package function

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/alexellis/derek/auth"
	"github.com/google/go-github/github"
	"github.com/openfaas/openfaas-cloud/sdk"
)

const (
	defaultPrivateKeyName = "private_key.pem"
)

var (
	token        = ""
	serviceValue = ""
)

// Handle a serverless request
func Handle(req []byte) string {

	status, marshalerr := sdk.UnmarshalStatus(req)
	if marshalerr != nil {
		log.Fatal("failed to parse status json, error: ", marshalerr.Error)
	}

	if len(status.CommitStatuses) == 0 {
		log.Fatal("failed commit statuses are empty: ", status.CommitStatuses)
	}

	serviceValue = status.EventInfo.Owner + "-" + status.EventInfo.Repository

	// use auth token if provided
	if status.AuthToken != "" && sdk.ValidToken(status.AuthToken) {
		token = status.AuthToken
		log.Printf("reusing provided auth token\n")
	} else {
		var tokenErr error
		// NOTE: currently vendored derek auth package doesn't take the private key as input;
		// but expect it to be present at : "/run/secrets/derek-private-key"
		// as docker /secrets dir has limited permission we are bound to use secret named
		// as "derek-private-key"
		// the below lines should  be uncommented once the package is updated in derek project
		// privateKeyPath := getPrivateKey()
		// token, tokenErr = auth.MakeAccessTokenForInstallation(os.Getenv("github_app_id"),
		//      event.installationID, privateKeyPath)
		token, tokenErr = auth.MakeAccessTokenForInstallation(os.Getenv("github_app_id"), status.EventInfo.InstallationID)
		if tokenErr != nil {
			log.Fatalf("failed to report status %v, error: %s\n", status, tokenErr.Error())
		}

		if token == "" {
			log.Fatalf("failed to report status %v, error: authentication failed Invalid token\n", status)
		}

		log.Printf("auth token is created")
	}

	for _, commitStatus := range status.CommitStatuses {
		err := ReportStatus(commitStatus.Status, commitStatus.Description, commitStatus.Context, &status.EventInfo)
		if err != nil {
			log.Fatalf("failed to report status %v, error: %s", status, err.Error())
		}
	}

	return token
}

func buildPublicStatusURL(status string, statusContext string, event *sdk.Event) string {
	url := event.URL

	if status == "success" {
		publicURL := os.Getenv("gateway_public_url")
		gatewayPrettyURL := os.Getenv("gateway_pretty_url")
		if statusContext != sdk.Stack {
			if len(gatewayPrettyURL) > 0 {
				// https://user.get-faas.com/function
				url = strings.Replace(gatewayPrettyURL, "user", event.Owner, 1)
				url = strings.Replace(url, "function", event.Service, 1)
			} else if len(publicURL) > 0 {
				if strings.HasSuffix(publicURL, "/") == false {
					publicURL = publicURL + "/"
				}
				// for success status if gateway's public url id set the deployed
				// function url is used in the commit status
				serviceValue := fmt.Sprintf("%s-%s", event.Owner, event.Service)
				url = publicURL + "function/" + serviceValue
			}
		} else { // For context Stack on success the gateway url is used
			if len(gatewayPrettyURL) > 0 {
				// https://user.get-faas.com/function
				url = strings.Replace(gatewayPrettyURL, "user", event.Owner, 1)
				url = strings.Replace(url, "function", "", 1)
			} else if len(publicURL) > 0 {
				if strings.HasSuffix(publicURL, "/") == false {
					publicURL = publicURL + "/"
				}
				url = publicURL
			}
		}
	}

	return url
}

func ReportStatus(status string, desc string, statusContext string, event *sdk.Event) error {

	ctx := context.Background()

	url := buildPublicStatusURL(status, statusContext, event)

	repoStatus := buildStatus(status, desc, statusContext, url)

	log.Printf("Status: %s, Context: %s, Service: %s, GitHub AppID: %d, Repo: %s, Owner: %s", status, statusContext, serviceValue, event.InstallationID, event.Repository, event.Owner)

	client := auth.MakeClient(ctx, token)

	_, _, apiErr := client.Repositories.CreateStatus(ctx, event.Owner, event.Repository, event.Sha, repoStatus)
	if apiErr != nil {
		return fmt.Errorf("failed to report status %v, error: %s\n", repoStatus, apiErr.Error())
	}

	return nil
}

func getPrivateKey() string {
	// we are taking the secrets name from the env, by default it is fixed
	// to private_key.pem.
	// Although user can make the secret with a specific name and provide
	// it in the stack.yaml and also specify the secret name in github.yml
	privateKeyName := os.Getenv("private_key")
	if privateKeyName == "" {
		privateKeyName = defaultPrivateKeyName
	}
	privateKeyPath := "/run/secrets/" + privateKeyName
	return privateKeyPath
}

func buildStatus(status string, desc string, context string, url string) *github.RepoStatus {
	return &github.RepoStatus{State: &status, TargetURL: &url, Description: &desc, Context: &context}
}

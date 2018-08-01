package function

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexellis/derek/auth"
	"github.com/google/go-github/github"
	"github.com/openfaas/openfaas-cloud/sdk"
)

const (
	defaultPrivateKeyName  = "private-key"
	defaultSecretMountPath = "/var/openfaas/secrets"
)

var (
	token        = ""
	serviceValue = ""
)

// Handle a serverless request
func Handle(req []byte) string {

	status, marshalerr := sdk.UnmarshalStatus(req)
	if marshalerr != nil {
		log.Fatal("failed to parse status request json, error: ", marshalerr.Error())
	}

	if len(status.CommitStatuses) == 0 {
		log.Fatal("failed commit statuses are empty: ", status.CommitStatuses)
	}

	serviceValue = status.EventInfo.Owner + "-" + status.EventInfo.Repository

	// use auth token if provided
	if status.AuthToken != "" && sdk.ValidToken(status.AuthToken) {
		token = status.AuthToken
		log.Printf("reusing provided auth token")
	} else {
		var tokenErr error
		privateKeyPath := getPrivateKey()
		token, tokenErr = auth.MakeAccessTokenForInstallation(os.Getenv("github_app_id"), status.EventInfo.InstallationID, privateKeyPath)
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

	// return auth token so that it can be reused form a same function
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
	// Private key name can be different from the default 'private-key'
	// When providing a different name in the stack.yaml, user need to specify the name
	// in github.yml as `private_key: <user_private_key>`
	privateKeyName := os.Getenv("private_key")
	if privateKeyName == "" {
		privateKeyName = defaultPrivateKeyName
	}
	secretMountPath := os.Getenv("secret_mount_path")
	if secretMountPath == "" {
		secretMountPath = defaultSecretMountPath
	}
	privateKeyPath := filepath.Join(secretMountPath, privateKeyName)
	return privateKeyPath
}

func buildStatus(status string, desc string, context string, url string) *github.RepoStatus {
	return &github.RepoStatus{State: &status, TargetURL: &url, Description: &desc, Context: &context}
}

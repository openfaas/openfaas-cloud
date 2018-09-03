package function

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/alexellis/derek/auth"
	"github.com/alexellis/derek/factory"
	"github.com/alexellis/hmac"
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

	if hmacEnabled() {

		key, keyErr := sdk.ReadSecret("payload-secret")
		if keyErr != nil {
			fmt.Fprintf(os.Stderr, keyErr.Error())
			os.Exit(-1)
		}

		digest := os.Getenv("Http_X_Cloud_Signature")

		validated := hmac.Validate(req, digest, key)

		if validated != nil {
			fmt.Fprintf(os.Stderr, validated.Error())
			os.Exit(-1)
		}
		fmt.Fprintf(os.Stderr, "hash for HMAC validated successfully\n")
	}

	status, marshalErr := sdk.UnmarshalStatus(req)
	if marshalErr != nil {
		log.Fatal("failed to parse status request json, error: ", marshalErr.Error())
	}

	if len(status.CommitStatuses) == 0 {
		log.Fatal("failed commit statuses are empty: ", status.CommitStatuses)
	}

	serviceValue = status.EventInfo.Owner + "-" + status.EventInfo.Repository

	// use auth token if provided
	if status.AuthToken != sdk.EmptyAuthToken && sdk.ValidToken(status.AuthToken) {
		token = status.AuthToken
		log.Printf("reusing provided auth token")
	} else {
		var tokenErr error
		privateKeyPath := sdk.GetPrivateKeyPath()
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
		err := reportStatus(commitStatus.Status, commitStatus.Description, commitStatus.Context, &status.EventInfo)
		if err != nil {
			log.Fatalf("failed to report status %v, error: %s", status, err.Error())
		}
	}

	if val, exists := os.LookupEnv("debug_token"); exists {
		if val == "true" {
			log.Printf("Token: %s for Installation: %d", token, status.EventInfo.InstallationID)
		}
	}

	// marshal token
	token = sdk.MarshalToken(token)

	// return auth token so that it can be reused form a same function
	return token
}

func buildPublicStatusURL(status string, statusContext string, event *sdk.Event) string {
	url := event.URL

	if status == "success" {
		publicURL := os.Getenv("gateway_public_url")
		gatewayPrettyURL := os.Getenv("gateway_pretty_url")
		if statusContext != sdk.StackContext {
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
				serviceValue := sdk.ServiceName(event.Owner, event.Service)
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

func reportStatus(status string, desc string, statusContext string, event *sdk.Event) error {

	ctx := context.Background()

	url := buildPublicStatusURL(status, statusContext, event)

	repoStatus := buildStatus(status, desc, statusContext, url)

	log.Printf("Status: %s, Context: %s, Service: %s, GitHub AppID: %d, Repo: %s, Owner: %s", status, statusContext, serviceValue, event.InstallationID, event.Repository, event.Owner)

	client := factory.MakeClient(ctx, token)

	_, _, apiErr := client.Repositories.CreateStatus(ctx, event.Owner, event.Repository, event.SHA, repoStatus)
	if apiErr != nil {
		return fmt.Errorf("failed to report status %v, error: %s", repoStatus, apiErr.Error())
	}

	return nil
}

func hmacEnabled() bool {
	return os.Getenv("validate_hmac") == "1" || os.Getenv("validate_hmac") == "true"
}

func buildStatus(status string, desc string, context string, url string) *github.RepoStatus {
	return &github.RepoStatus{State: &status, TargetURL: &url, Description: &desc, Context: &context}
}

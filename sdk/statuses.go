package sdk

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/alexellis/derek/auth"
	"github.com/google/go-github/github"
)

const (
	defaultPrivateKeyName = "private_key.pem"
)

var (
	token = ""
)

type EventInfo struct {
	Service        string
	Owner          string
	Repository     string
	Sha            string
	URL            string
	PublicURL      string
	InstallationID int
}

func ReportStatus(status string, desc string, statusContext string, event *EventInfo) {

	serviceValue := fmt.Sprintf("%s-%s", event.Owner, event.Service)

	url := event.URL
	if status == "success" {
		// for success status if gateway's public url id set the deployed
		// function url is used in the commit status
		if event.PublicURL != "" {
			url = event.PublicURL + "function/" + serviceValue
		}
	}

	ctx := context.Background()

	repoStatus := buildStatus(status, desc, statusContext, url)

	log.Printf("Status: %s, Context: %s, Service: %s, GitHub AppID: %d, Repo: %s, Owner: %s", status, statusContext, serviceValue, event.InstallationID, event.Repository, event.Owner)

	if token == "" {
		var tokenErr error
		// NOTE: currently vendored derek auth package doesn't take the private key as input;
		// but expect it to be present at : "/run/secrets/derek-private-key"
		// as docker /secrets dir has limited permission we are bound to use secret named
		// as "derek-private-key"
		// the below lines should  be uncommented once the package is updated in derek project
		// privateKeyPath := getPrivateKey()
		// token, tokenErr = auth.MakeAccessTokenForInstallation(os.Getenv("github_app_id"),
		//      event.installationID, privateKeyPath)

		token, tokenErr = auth.MakeAccessTokenForInstallation(os.Getenv("github_app_id"), event.InstallationID)
		if tokenErr != nil {
			fmt.Printf("failed to report status %v, error: %s\n", repoStatus, tokenErr.Error())
			return
		}

		if token == "" {
			fmt.Printf("failed to report status %v, error: authentication failed Invalid token\n", repoStatus)
			return
		}
	}

	client := auth.MakeClient(ctx, token)

	_, _, apiErr := client.Repositories.CreateStatus(ctx, event.Owner, event.Repository, event.Sha, repoStatus)
	if apiErr != nil {
		fmt.Printf("failed to report status %v, error: %s\n", repoStatus, apiErr.Error())
		return
	}
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

package function

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/alexellis/hmac"
)

// Handle a serverless request
func Handle(req []byte) string {
	event := os.Getenv("Http_X_Github_Event")

	if event == "installation_repositories" {
		xHubSignature := os.Getenv("Http_X_Hub_Signature")

		shouldValidate := os.Getenv("validate_hmac")
		if len(shouldValidate) > 0 && (shouldValidate == "1" || shouldValidate == "true") {
			validateErr := hmac.Validate(req, xHubSignature, os.Getenv("github_webhook_secret"))
			if validateErr != nil {
				log.Fatal(validateErr)
			}
		}

		event := InstallationRepositoriesEvent{}
		err := json.Unmarshal(req, &event)
		if err != nil {
			return err.Error()
		}

		switch event.Action {
		case "removed":
			garbageRequests := []GarbageRequest{}
			for _, repo := range event.Installation.RepositoriesRemoved {
				fmt.Printf("Need to remove: %s.\n", repo.FullName)

				garbageRequests = append(garbageRequests,
					GarbageRequest{
						Functions: []string{},
						Owner:     event.Installation.Account.Login,
						Repo:      repo.Name,
					},
				)
			}
			garbageCollect(garbageRequests)
			break
		}

	}

	return fmt.Sprintf("Message received")
}

func garbageCollect(garbageRequests []GarbageRequest) error {
	client := http.Client{}
	for _, garbageRequest := range garbageRequests {

		body, _ := json.Marshal(garbageRequest)
		bodyReader := bytes.NewReader(body)
		req, _ := http.NewRequest(http.MethodPost, "http://gateway:8080/function/garbage-collect", bodyReader)
		res, err := client.Do(req)
		if err != nil {
			return err
		}
		if res.Body != nil {
			defer res.Body.Close()
		}
		if res.StatusCode != http.StatusOK {
			resBody, _ := ioutil.ReadAll(res.Body)
			fmt.Printf("Error in garbageCollect: %s\n", resBody)
		}
	}
	return nil
}

type GarbageRequest struct {
	Functions []string
	Repo      string
	Owner     string
}

type InstallationRepositoriesEvent struct {
	Action       string `json:"action"`
	Installation struct {
		Account struct {
			Login string
		}
		RepositoriesRemoved []Installation `json:"repositories_removed"`
		RepositoriesAdded   []Installation `json:"repositories_added"`
	}
}

type Installation struct {
	Name     string
	FullName string `json:"full_name"`
}

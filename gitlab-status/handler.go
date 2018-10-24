package function

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/alexellis/hmac"
	"github.com/openfaas/openfaas-cloud/sdk"
)

// Handle a serverless request
func Handle(req []byte) string {
	validateError := validateRequest(req)
	if validateError != nil {
		log.Fatal(validateError)
	}

	status, statusErr := sdk.UnmarshalStatus(req)
	if statusErr != nil {
		return statusErr.Error()
	}

	token, tokenErr := sdk.ReadSecret("gitlab-api-token")
	if tokenErr != nil {
		return tokenErr.Error()
	}

	url, urlErr := gitLabURLBuilder(status.EventInfo.URL, status.EventInfo.SHA, status.EventInfo.InstallationID)
	if urlErr != nil {
		log.Printf("error while building URL: %s", urlErr.Error())
	}

	for _, commitStatus := range status.CommitStatuses {
		reportErr := sendReport(url, token, commitStatus.Status, commitStatus.Description, commitStatus.Context)
		if reportErr != nil {
			log.Fatalf("failed to report status %v, error: %s", status, reportErr.Error())
		}
	}

	return ""
}

func gitLabURLBuilder(eventURL, SHA string, id int) (string, error) {
	if eventURL == "" || SHA == "" {
		return "", fmt.Errorf("eventURL or SHA are empty")
	}
	parsedURL, parseErr := url.Parse(eventURL)
	if parseErr != nil {
		return "", fmt.Errorf("Error while parsing eventURL: %s", parseErr.Error())
	}
	return fmt.Sprintf("%s://%s/api/v4/projects/%d/statuses/%s", parsedURL.Scheme, parsedURL.Host, id, SHA), nil
}

func appendParameters(URL string, state string, desc string, context string) (string, error) {
	var theURL *url.URL

	theURL, urlErr := url.Parse(URL)
	if urlErr != nil {
		return "", fmt.Errorf("error while appending parameters to url: %s", urlErr.Error())
	}

	if state == "failure" {
		state = "failed"
	}

	parameters := url.Values{}
	parameters.Add("state", state)
	parameters.Add("description", desc)
	parameters.Add("context", context)
	theURL.RawQuery = parameters.Encode()

	return theURL.String(), nil

}

func sendReport(URL string, token string, state string, desc string, context string) error {
	fullURL, fullURLErr := appendParameters(URL, state, desc, context)
	if fullURLErr != nil {
		return fullURLErr
	}
	var b io.Reader
	req, reqErr := http.NewRequest("POST", fullURL, b)
	if reqErr != nil {
		return reqErr
	}
	req.Header.Set("PRIVATE-TOKEN", token)
	client := &http.Client{}
	resp, clientErr := client.Do(req)
	if clientErr != nil {
		return clientErr
	}
	resp.Body.Close()

	return nil
}

func validateRequest(req []byte) (err error) {
	payloadSecret, err := sdk.ReadSecret("payload-secret")

	if err != nil {
		return fmt.Errorf("couldn't get payload-secret: %t", err)
	}

	xCloudSignature := os.Getenv("Http_X_Cloud_Signature")

	err = hmac.Validate(req, xCloudSignature, payloadSecret)

	if err != nil {
		return err
	}

	return nil
}

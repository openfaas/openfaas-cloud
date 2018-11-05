package function

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/alexellis/hmac"
	"github.com/openfaas/openfaas-cloud/sdk"
)

const (
	SystemSubDomain = "system"
	logsURL         = "logsURL"
	endpointURL     = "endpointURL"
	dashboardURL    = "dashboardURL"
)

// Handle reports the building process of the
// function and the function stack to GitLab by
// sending commit statuses on pending, success, failure
func Handle(req []byte) string {
	if validateError := validateRequest(req); validateError != nil {
		log.Fatal(validateError)
	}

	status, statusErr := sdk.UnmarshalStatus(req)
	if statusErr != nil {
		return fmt.Sprintf("error while un-marshaling status from request: %s", statusErr.Error())
	}

	token, tokenErr := sdk.ReadSecret("gitlab-api-token")
	if tokenErr != nil {
		return fmt.Sprintf("error while reading gitlab-api-token: %s", tokenErr.Error())
	}

	url, urlErr := gitLabURLBuilder(status.EventInfo.URL, status.EventInfo.SHA, status.EventInfo.InstallationID)
	if urlErr != nil {
		log.Printf("error while building base URL to the API: %s", urlErr.Error())
	}

	gatewayURL := os.Getenv("gateway_public_url")
	if gatewayURL == "" {
		log.Printf("empty gateway_public_url unable to construct logs and endpoint URLs")
	}

	endpointURL, targetURLerr := getTargetURLs(gatewayURL, &status.EventInfo)
	if targetURLerr != nil {
		log.Printf("error while obtaining available target URLs: %s", targetURLerr.Error())
	}

	for _, commitStatus := range status.CommitStatuses {
		targetURL := setTargetURL(commitStatus.Status, commitStatus.Context,
			status.EventInfo.Service, endpointURL)

		reportErr := sendReport(url, token, commitStatus.Status,
			commitStatus.Description, commitStatus.Context, targetURL)
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
		return "", fmt.Errorf("error while parsing eventURL: %s", parseErr.Error())
	}
	return fmt.Sprintf("%s://%s/api/v4/projects/%d/statuses/%s", parsedURL.Scheme, parsedURL.Host, id, SHA), nil
}

func appendParameters(URL, state, desc, context, targetURL string) (string, error) {
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
	parameters.Add("target_url", targetURL)
	theURL.RawQuery = parameters.Encode()

	return theURL.String(), nil

}

func sendReport(URL, token, state, desc, context, targetURL string) error {
	fullURL, fullURLErr := appendParameters(URL, state, desc, context, targetURL)
	if fullURLErr != nil {
		return fmt.Errorf("error while appending parameters to URL: %s", fullURLErr)
	}

	req, reqErr := http.NewRequest(http.MethodPost, fullURL, nil)
	if reqErr != nil {
		return fmt.Errorf("error while creating request to GitLab API: %s", reqErr.Error())
	}
	req.Header.Set("PRIVATE-TOKEN", token)

	resp, clientErr := http.DefaultClient.Do(req)
	if clientErr != nil {
		return fmt.Errorf("error while sending request to GitLab API: %s", clientErr.Error())
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	return nil
}

func validateRequest(req []byte) (err error) {
	payloadSecret, err := sdk.ReadSecret("payload-secret")
	if err != nil {
		return fmt.Errorf("couldn't get payload-secret: %t", err)
	}

	xCloudSignature := os.Getenv("Http_X_Cloud_Signature")

	if err = hmac.Validate(req, xCloudSignature, payloadSecret); err != nil {
		return err
	}

	return nil
}

func setTargetURL(functionStatus, functionContext, eventService string, availableURL Endpoints) string {
	targetURL := availableURL.FunctionEndpointURL
	if functionStatus != sdk.StatusSuccess {
		targetURL = availableURL.LogsURL
	}
	if functionContext == sdk.StackContext &&
		functionContext != eventService {
		targetURL = availableURL.DashboardURL
	}
	return targetURL
}

func getTargetURLs(gatewayURL string, event *sdk.Event) (Endpoints, error) {
	var availableTargetURL Endpoints
	var targetURLerr error
	availableTargetURL.LogsURL, targetURLerr = sdk.FormatLogsURL(gatewayURL, event)
	if targetURLerr != nil {
		return availableTargetURL, fmt.Errorf("error while formatting URL for logs: %s", targetURLerr.Error())
	}

	availableTargetURL.FunctionEndpointURL, targetURLerr = sdk.FormatEndpointURL(gatewayURL, event)
	if targetURLerr != nil {
		return availableTargetURL, fmt.Errorf("error while formatting endpoint URL: %s", targetURLerr.Error())
	}

	availableTargetURL.DashboardURL, targetURLerr = sdk.FormatDashboardURL(gatewayURL, event)
	if targetURLerr != nil {
		return availableTargetURL, fmt.Errorf("the dashboard url failed: %s", targetURLerr.Error())
	}

	return availableTargetURL, nil
}

type Endpoints struct {
	LogsURL             string
	FunctionEndpointURL string
	DashboardURL        string
}

package function

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/alexellis/hmac"
	"github.com/openfaas/openfaas-cloud/sdk"
)

const Source = "gh-push"

// Handle a serverless request
func Handle(req []byte) string {

	event := os.Getenv("Http_X_Github_Event")
	if event != "push" {

		auditEvent := sdk.AuditEvent{
			Message: "bad event: " + event,
			Source:  Source,
		}

		sdk.PostAudit(auditEvent)

		return fmt.Sprintf("gh-push cannot handle event: %s", event)
	}

	xHubSignature := os.Getenv("Http_X_Hub_Signature")

	shouldValidate := readBool("validate_hmac")
	if shouldValidate {
		validateErr := hmac.Validate(req, xHubSignature, os.Getenv("github_webhook_secret"))
		if validateErr != nil {
			log.Fatal(validateErr)
		}
	}

	pushEvent := sdk.PushEvent{}
	err := json.Unmarshal(req, &pushEvent)
	if err != nil {
		return err.Error()
	}

	var found bool

	if readBool("validate_customers") {

		customersURL := os.Getenv("customers_url")

		customers, getErr := getCustomers(customersURL)
		if getErr != nil {
			return getErr.Error()
		}

		for _, customer := range customers {
			if customer == pushEvent.Repository.Owner.Login {
				found = true
			}
		}
		if !found {

			auditEvent := sdk.AuditEvent{
				Message: "Customer not found",
				Owner:   pushEvent.Repository.Owner.Login,
				Repo:    pushEvent.Repository.Name,
				Source:  Source,
			}

			sdk.PostAudit(auditEvent)

			return fmt.Sprintf("Customer: %s not found in CUSTOMERS file via %s", pushEvent.Repository.Owner.Login, customersURL)
		}
	}

	if pushEvent.Ref != "refs/heads/master" {
		msg := "refusing to build non-master branch: " + pushEvent.Ref
		auditEvent := sdk.AuditEvent{
			Message: msg,
			Owner:   pushEvent.Repository.Owner.Login,
			Repo:    pushEvent.Repository.Name,
			Source:  Source,
		}

		sdk.PostAudit(auditEvent)
		return msg
	}

	statusCode, postErr := postEvent(pushEvent)
	if postErr != nil {
		return postErr.Error()
	}

	auditEvent := sdk.AuditEvent{
		Message: "Git-tar invoked",
		Owner:   pushEvent.Repository.Owner.Login,
		Repo:    pushEvent.Repository.Name,
		Source:  Source,
	}

	sdk.PostAudit(auditEvent)

	return fmt.Sprintf("Push - %v, git-tar status: %d\n", pushEvent, statusCode)
}

// getCustomers reads a list of customers separated by new lines
// who are valid users of OpenFaaS cloud
func getCustomers(customerURL string) ([]string, error) {
	customers := []string{}

	c := http.Client{}

	httpReq, _ := http.NewRequest(http.MethodGet, customerURL, nil)
	res, reqErr := c.Do(httpReq)

	if reqErr != nil {
		return customers, reqErr
	}

	if res.Body != nil {
		defer res.Body.Close()

		pageBody, _ := ioutil.ReadAll(res.Body)
		customers = strings.Split(string(pageBody), "\n")
	}

	return customers, nil
}

func postEvent(pushEvent sdk.PushEvent) (int, error) {
	gatewayURL := os.Getenv("gateway_url")

	body, _ := json.Marshal(pushEvent)

	c := http.Client{}
	bodyReader := bytes.NewBuffer(body)
	httpReq, _ := http.NewRequest(http.MethodPost, gatewayURL+"async-function/git-tar", bodyReader)
	res, reqErr := c.Do(httpReq)

	if reqErr != nil {
		return http.StatusServiceUnavailable, reqErr
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	return res.StatusCode, nil
}

func init() {

}

func readBool(key string) bool {
	if val, exists := os.LookupEnv(key); exists {
		return val == "true" || val == "1"
	}
	return false
}

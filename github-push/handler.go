package function

import (
	"bytes"
	"encoding/hex"
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

// Source name for this function when auditing
const Source = "github-push"

//SCM identifier
const SCM = "github"

var audit sdk.Audit

// Handle processes the push event from the "github-event" function
func Handle(req []byte) string {

	if audit == nil {
		audit = sdk.AuditLogger{}
	}

	event := os.Getenv("Http_X_Github_Event")
	if event != "push" {

		auditEvent := sdk.AuditEvent{
			Message: "bad event: " + event,
			Source:  Source,
		}
		audit.Post(auditEvent)

		return fmt.Sprintf("%s cannot handle event: %s", Source, event)
	}

	xHubSignature := os.Getenv("Http_X_Hub_Signature")

	shouldValidate := readBool("validate_hmac")
	if shouldValidate {
		webhookSecretKey, secretErr := sdk.ReadSecret("github-webhook-secret")
		if secretErr != nil {
			return secretErr.Error()
		}

		validateErr := hmac.Validate(req, xHubSignature, webhookSecretKey)
		if validateErr != nil {
			log.Fatal(validateErr)
		}
	}

	pushEvent := sdk.PushEvent{}
	err := json.Unmarshal(req, &pushEvent)

	if err != nil {
		return err.Error()
	}

	pushEvent.SCM = SCM

	var found bool

	if readBool("validate_customers") {

		customersURL := os.Getenv("customers_url")

		customers, getErr := getCustomers(customersURL)
		if getErr != nil {
			return getErr.Error()
		}

		found = validCustomer(customers, pushEvent.Repository.Owner.Login)

		if !found {

			auditEvent := sdk.AuditEvent{
				Message: "Customer not found",
				Owner:   pushEvent.Repository.Owner.Login,
				Repo:    pushEvent.Repository.Name,
				Source:  Source,
			}

			audit.Post(auditEvent)

			return fmt.Sprintf("Customer: %s not found in CUSTOMERS file via %s", pushEvent.Repository.Owner.Login, customersURL)
		}
	}

	eventInfo := sdk.BuildEventFromPushEvent(pushEvent)
	status := sdk.BuildStatus(eventInfo, sdk.EmptyAuthToken)

	if len(pushEvent.Ref) == 0 ||
		pushEvent.Ref != "refs/heads/master" {
		msg := "refusing to build non-master branch: " + pushEvent.Ref
		auditEvent := sdk.AuditEvent{
			Message: msg,
			Owner:   pushEvent.Repository.Owner.Login,
			Repo:    pushEvent.Repository.Name,
			Source:  Source,
		}

		audit.Post(auditEvent)

		status.AddStatus(sdk.StatusFailure, msg, sdk.StackContext)
		reportGitHubStatus(status)
		return msg
	}

	serviceValue := sdk.FormatServiceName(pushEvent.Repository.Owner.Login, pushEvent.Repository.Name)

	status.AddStatus(sdk.StatusPending, fmt.Sprintf("%s stack deploy is in progress", serviceValue), sdk.StackContext)
	reportGitHubStatus(status)

	statusCode, postErr := postEvent(pushEvent)
	if postErr != nil {
		status.AddStatus(sdk.StatusFailure, postErr.Error(), sdk.StackContext)
		reportGitHubStatus(status)
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

	if len(customerURL) == 0 {
		return nil, fmt.Errorf("customerURL was nil")
	}

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

		for i, c := range customers {
			customers[i] = strings.ToLower(c)
		}

	}

	return customers, nil
}

func postEvent(pushEvent sdk.PushEvent) (int, error) {
	gatewayURL := os.Getenv("gateway_url")

	payloadSecret, err := sdk.ReadSecret("payload-secret")
	if err != nil {
		return http.StatusUnauthorized, err
	}

	body, _ := json.Marshal(pushEvent)

	c := http.Client{}
	bodyReader := bytes.NewBuffer(body)
	httpReq, _ := http.NewRequest(http.MethodPost, gatewayURL+"async-function/git-tar", bodyReader)

	digest := hmac.Sign(body, []byte(payloadSecret))
	httpReq.Header.Add(sdk.CloudSignatureHeader, "sha1="+hex.EncodeToString(digest))

	res, reqErr := c.Do(httpReq)

	if reqErr != nil {
		return http.StatusServiceUnavailable, reqErr
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	return res.StatusCode, nil
}

func readBool(key string) bool {
	if val, exists := os.LookupEnv(key); exists {
		return val == "true" || val == "1"
	}
	return false
}

func enableStatusReporting() bool {
	return os.Getenv("report_status") == "true"
}

func reportGitHubStatus(status *sdk.Status) {

	if !enableStatusReporting() {
		return
	}

	hmacKey, keyErr := sdk.ReadSecret("payload-secret")
	if keyErr != nil {
		log.Printf("failed to load hmac key for status, error " + keyErr.Error())
		return
	}

	gatewayURL := os.Getenv("gateway_url")

	_, reportErr := status.Report(gatewayURL, hmacKey)
	if reportErr != nil {
		log.Printf("failed to report status, error: %s", reportErr.Error())
	}
}

func init() {

}

func validCustomer(customers []string, owner string) bool {
	found := false
	for _, customer := range customers {
		if len(customer) > 0 &&
			strings.EqualFold(customer, owner) {
			found = true
			break
		}
	}

	return found
}

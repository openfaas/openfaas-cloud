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
)

const Source = "gh-push"

// Handle a serverless request
func Handle(req []byte) string {

	event := os.Getenv("Http_X_Github_Event")

	if event != "push" {

		auditEvent := AuditEvent{
			Message: "bad event: " + event,
			Source:  Source,
		}

		postAudit(auditEvent)

		return fmt.Sprintf("gh-push cannot handle event: %s", event)
	}

	xHubSignature := os.Getenv("Http_X_Hub_Signature")

	if readFlag("validate_hmac") == true {
		validateErr := hmac.Validate(req, xHubSignature, os.Getenv("github_webhook_secret"))
		if validateErr != nil {
			log.Fatal(validateErr)
		}
	}

	pushEvent := PushEvent{}
	err := json.Unmarshal(req, &pushEvent)
	if err != nil {
		return err.Error()
	}

	var found bool

	if readFlag("validate_customers") {
		customersURL := os.Getenv("customers_url")

		customers, getErr := getCustomers(customersURL)
		if getErr != nil {
			return getErr.Error()
		}

		found := false
		for _, customer := range customers {
			if customer == pushEvent.Repository.Owner.Login {
				found = true
			}
		}
	} else {
		found = true
	}

	if !found {

		auditEvent := AuditEvent{
			Message: "Customer not found",
			Owner:   pushEvent.Repository.Owner.Login,
			Repo:    pushEvent.Repository.Name,
			Source:  Source,
		}

		postAudit(auditEvent)

		return fmt.Sprintf("Customer: %s not found in CUSTOMERS file via %s", pushEvent.Repository.Owner.Login, customersURL)
	}

	statusCode, postErr := postEvent(pushEvent)
	if postErr != nil {
		return postErr.Error()
	}

	auditEvent := AuditEvent{
		Message: "Git-tar invoked",
		Owner:   pushEvent.Repository.Owner.Login,
		Repo:    pushEvent.Repository.Name,
		Source:  Source,
	}

	postAudit(auditEvent)

	return fmt.Sprintf("Push - %s, git-tar status: %d\n", pushEvent, statusCode)
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

func postEvent(pushEvent PushEvent) (int, error) {
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

// PushEvent as received from GitHub
type PushEvent struct {
	Repository struct {
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		CloneURL string `json:"clone_url"`
		Owner    struct {
			Login string `json:"login"`
			Email string `json:"email"`
		} `json:"owner"`
	}
	AfterCommitID string `json:"after"`
}

func Init() {

}

func readFlag(key string) bool {
	if val, exists := os.LookupEnv(key); val == "true" || val == "1" {
		return true
	}
	return false
}

// Move method / struct to separate package

func postAudit(auditEvent AuditEvent) {
	c := http.Client{}
	bytesOut, _ := json.Marshal(&auditEvent)
	reader := bytes.NewBuffer(bytesOut)

	req, _ := http.NewRequest(http.MethodPost, os.Getenv("audit_url"), reader)

	res, err := c.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
}

type AuditEvent struct {
	Source  string
	Message string
	Owner   string
	Repo    string
}

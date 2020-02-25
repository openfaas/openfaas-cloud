package function

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/alexellis/hmac"
	"github.com/openfaas/openfaas-cloud/sdk"
)

// Source name for this function when auditing
const Source = "github-event"

var audit sdk.Audit

type GarbageRequest struct {
	Functions []string `json:"functions"`
	Repo      string   `json:"repo"`
	Owner     string   `json:"owner"`
}

type InstallationRepositoriesEvent struct {
	Action       string `json:"action"`
	Installation struct {
		Account struct {
			Login string
		}
	} `json:"installation"`
	RepositoriesRemoved []Installation `json:"repositories_removed"`
	RepositoriesAdded   []Installation `json:"repositories_added"`
	Repositories        []Installation `json:"repositories"`
}

type Installation struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
}

// Handle receives events from the GitHub app and checks the origin via
// HMAC. Valid events are push or installation events.
func Handle(req []byte) string {
	customersPath := os.Getenv("customers_path")
	customersURL := os.Getenv("customers_url")

	customers := sdk.NewCustomers(customersPath, customersURL)
	customers.Fetch()

	queryVal := os.Getenv("Http_Query")
	if values, err := url.ParseQuery(queryVal); err == nil {
		setupAction := values.Get("setup_action")
		if setupAction == "install" {
			return "Installation completed, please return to the installation guide."
		}
	}

	if audit == nil {
		audit = sdk.AuditLogger{}
	}

	eventHeader := os.Getenv("Http_X_Github_Event")
	xHubSignature := os.Getenv("Http_X_Hub_Signature")

	if eventHeader != "push" &&
		eventHeader != "installation_repositories" &&
		eventHeader != "integration_installation" &&
		eventHeader != "installation" {

		auditEvent := sdk.AuditEvent{
			Message: "bad event: " + eventHeader,
			Source:  Source,
		}

		sdk.PostAudit(auditEvent)

		return fmt.Sprintf("%s cannot handle event: %s", Source, eventHeader)
	}

	customer := sdk.PushEvent{}
	unmarshalErr := json.Unmarshal(req, &customer)
	if unmarshalErr != nil {
		return fmt.Sprintf("Error while un-marshaling customers: %s, value: %s",
			unmarshalErr.Error(),
			string(req))
	}

	if eventHeader == "push" {
		if sdk.ValidateCustomers() {
			err := validateCustomers(&customer, customers)
			if err != nil {
				return err.Error()
			}
		}

		if sdk.HmacEnabled() {
			webhookSecretKey, secretErr := sdk.ReadSecret("github-webhook-secret")
			if secretErr != nil {
				return secretErr.Error()
			}

			validateErr := hmac.Validate(req, xHubSignature, webhookSecretKey)
			if validateErr != nil {
				log.Fatal(validateErr)
			}
		}

		headers := map[string]string{
			"X-Hub-Signature": xHubSignature,
			"X-GitHub-Event":  eventHeader,
			"Content-Type":    "application/json",
		}

		body, statusCode, err := forward(req, "github-push", headers)

		if statusCode == http.StatusOK {
			return fmt.Sprintf("Forwarded to function: %d, %s", statusCode, body)
		}

		if err != nil {
			return err.Error()
		}

		return body
	}

	if eventHeader == "installation" ||
		eventHeader == "installation_repositories" ||
		eventHeader == "integration_installation" {

		event := InstallationRepositoriesEvent{}
		err := json.Unmarshal(req, &event)
		if err != nil {
			return err.Error()
		}

		if sdk.ValidateCustomers() {
			customer := sdk.PushEvent{
				Repository: sdk.PushEventRepository{
					Owner: sdk.Owner{
						Login: event.Installation.Account.Login,
					},
				},
			}

			err := validateCustomers(&customer, customers)
			if err != nil {
				return err.Error()
			}
		}

		if sdk.HmacEnabled() {
			webhookSecretKey, secretErr := sdk.ReadSecret("github-webhook-secret")
			if secretErr != nil {
				return secretErr.Error()
			}

			validateErr := hmac.Validate(req, xHubSignature, webhookSecretKey)
			if validateErr != nil {
				log.Fatal(validateErr)
			}
		}

		fmt.Printf("event.Action: %s\n", event.Action)

		switch event.Action {
		case "created", "added":

			addedVal := ""
			if event.RepositoriesAdded != nil {
				for _, added := range event.RepositoriesAdded {
					addedVal += added.FullName + ", "
				}
			}
			if event.Repositories != nil {
				for _, added := range event.Repositories {
					addedVal += added.FullName + ", "
				}
			}

			auditEvent := sdk.AuditEvent{
				Message: event.Installation.Account.Login + " added repositories: " + addedVal,
				Source:  Source,
			}

			sdk.PostAudit(auditEvent)

		case "removed":
			garbageRequests := []GarbageRequest{}
			for _, repo := range event.RepositoriesRemoved {
				fmt.Printf("Need to remove: %s.\n", repo.FullName)

				garbageRequests = append(garbageRequests,
					GarbageRequest{
						Owner:     event.Installation.Account.Login,
						Repo:      repo.Name,
						Functions: []string{},
					},
				)
			}
			garbageCollect(garbageRequests)
			break
		case "deleted":
			garbageRequests := []GarbageRequest{}
			owner := event.Installation.Account.Login
			fmt.Printf("Need to remove all repos for owner: %s.\n", owner)

			garbageRequests = append(garbageRequests,
				GarbageRequest{
					Owner:     owner,
					Repo:      "*",
					Functions: []string{},
				},
			)

			garbageCollect(garbageRequests)

			break
		}

	}

	return fmt.Sprintf("Message received with event: %s", eventHeader)
}

func validateCustomers(pushEvent *sdk.PushEvent, customers *sdk.Customers) error {
	owner := pushEvent.Repository.Owner.Login

	notFound := fmt.Errorf("Customer: %q not found in customers ACL", owner)

	found1, err1 := customers.Get(owner)
	fmt.Println(owner, found1, err1)

	if found, err := customers.Get(owner); found == false || err != nil {

		if err != nil {
			log.Printf("Error getting customer: %s, %s", owner, err.Error())
		}

		auditEvent := sdk.AuditEvent{
			Message: "Customer not found",
			Owner:   owner,
			Source:  Source,
		}

		sdk.PostAudit(auditEvent)
		return notFound
	}

	return nil
}

func garbageCollect(garbageRequests []GarbageRequest) error {

	gatewayURL := os.Getenv("gateway_url")

	payloadSecret, err := sdk.ReadSecret("payload-secret")
	if err != nil {
		return err
	}

	for _, garbageRequest := range garbageRequests {

		body, _ := json.Marshal(garbageRequest)
		bodyReader := bytes.NewReader(body)
		req, _ := http.NewRequest(http.MethodPost, gatewayURL+"async-function/garbage-collect", bodyReader)

		digest := hmac.Sign(body, []byte(payloadSecret))
		req.Header.Add(sdk.CloudSignatureHeader, "sha1="+hex.EncodeToString(digest))

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		if res.Body != nil {
			defer res.Body.Close()
		}
		if res.StatusCode != http.StatusAccepted {
			log.Printf("Unexpected status code for function: `%s` - %d\n", garbageRequest.Repo, res.StatusCode)
			resBody, _ := ioutil.ReadAll(res.Body)
			fmt.Printf("Error in garbageCollect: %s\n", resBody)
		}
	}
	return nil
}

func forward(req []byte, function string, headers map[string]string) (string, int, error) {
	payloadSecret, err := sdk.ReadSecret("payload-secret")
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	bodyReader := bytes.NewBuffer(req)
	pushReq, _ := http.NewRequest(http.MethodPost, os.Getenv("gateway_url")+"function/"+function, bodyReader)
	digest := hmac.Sign(req, []byte(payloadSecret))
	pushReq.Header.Add(sdk.CloudSignatureHeader, "sha1="+hex.EncodeToString(digest))

	for k, v := range headers {
		pushReq.Header.Add(k, v)
	}

	res, err := http.DefaultClient.Do(pushReq)
	if err != nil {
		msg := "cannot post to " + function + ": " + err.Error()
		auditEvent := sdk.AuditEvent{
			Message: msg,
			Source:  Source,
		}
		sdk.PostAudit(auditEvent)
		return "", http.StatusInternalServerError, fmt.Errorf(msg)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}
	body, _ := ioutil.ReadAll(res.Body)

	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf(string(body))
	}

	return string(body), res.StatusCode, err
}

func readBool(key string) bool {
	if val, exists := os.LookupEnv(key); exists {
		return val == "true" || val == "1"
	}
	return false
}

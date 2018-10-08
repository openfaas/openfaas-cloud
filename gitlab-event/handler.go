package function

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/alexellis/hmac"
	"github.com/openfaas/openfaas-cloud/sdk"
)

// Source name for this function when auditing
const (
	Source              = "gitlab-event"
	EventSource         = "System Hook"
	PushEvent           = "push"
	ProjectUpdateEvent  = "project_update"
	ProjectDestroyEvent = "project_destroy"
)

var (
	supportedEvents = [...]string{PushEvent, ProjectUpdateEvent, ProjectDestroyEvent}
)

// Handle a serverless request
func Handle(req []byte) string {
	eventHeader := os.Getenv("Http_X_Gitlab_Event")
	xGitlabToken := os.Getenv("Http_X_Gitlab_Token")

	if eventHeader != EventSource {
		auditEvent := sdk.AuditEvent{
			Message: "required : " + EventSource,
			Source:  Source,
		}

		sdk.PostAudit(auditEvent)

		return fmt.Sprintf("%s: %s required cannot handle: %s", Source, EventSource, eventHeader)
	}

	eventName := PureEvent{}
	unmarshalErr := json.Unmarshal(req, &eventName)
	if unmarshalErr != nil {
		return fmt.Sprintf("error while un-marshaling event: %s", unmarshalErr.Error())
	}

	if !checkSupportedEvents(eventName.Event) {

		auditEvent := sdk.AuditEvent{
			Message: "bad event: " + eventName.Event,
			Source:  Source,
		}

		sdk.PostAudit(auditEvent)

		return fmt.Sprintf("%s cannot handle event: %s", Source, eventName.Event)
	}

	if readBool("validate_token") {
		tokenSecretKey, secretErr := sdk.ReadSecret("gitlab-webhook-secret")
		if secretErr != nil {
			return fmt.Sprintf("unable to load gitlab-webhook-secret: %s", secretErr.Error())
		}
		if xGitlabToken != tokenSecretKey {
			return fmt.Sprintf("The request token and the existing tokens mismatch")
		}
	}

	instance := os.Getenv("gitlab_instance")

	installationTag := "openfaas-cloud"
	if tag, exists := os.LookupEnv("installation_tag"); exists {
		installationTag = tag
	}

	switch eventName.Event {
	case PushEvent:
		eventInfo := sdk.GitLabPushEvent{}
		json.Unmarshal(req, &eventInfo)

		if readBool("validate_customers") {
			customersURL := os.Getenv("customers_url")
			customers, getErr := getCustomers(customersURL)
			if getErr != nil {
				return fmt.Sprintf("unable to read customers from %s error: %s", customersURL, getErr.Error())
			}
			if !validCustomer(customers, eventInfo.UserUsername) {
				auditEvent := sdk.AuditEvent{
					Message: "Customer not found",
					Owner:   eventInfo.UserUsername,
					Source:  Source,
				}
				sdk.PostAudit(auditEvent)
				return fmt.Sprintf("Customer: %s not found in CUSTOMERS file via %s", eventInfo.UserUsername, customersURL)
			}
		}

		installed, err := InstalledApp(eventInfo.GitLabProject.ID, instance, installationTag)
		if err != nil {
			return fmt.Sprintf("Error while trying to connect to GitLab API: %s", err.Error())
		}
		if installed {
			headers := map[string]string{
				"X-Gitlab-Token": xGitlabToken,
				"X-Gitlab-Event": eventHeader,
				"Content-Type":   "application/json",
			}

			body, statusCode, err := forward(req, "gitlab-push", headers)
			if err != nil {
				return fmt.Sprintf("error while forwarding to gitlab-push: %s", err.Error())
			}
			if statusCode == http.StatusOK {
				return fmt.Sprintf("Forwarded to function: `gitlab-push` status: %d response: %s", statusCode, body)
			}

			return body
		}
		return fmt.Sprintf("To install project on the openfaas-cloud instance add \"%s\" tag", installationTag)

	case ProjectUpdateEvent, ProjectDestroyEvent:
		eventInfo := GitLabProjectEvent{}
		unmarshalErr := json.Unmarshal(req, &eventInfo)
		if unmarshalErr != nil {
			return fmt.Sprintf("error while un-marshaling eventInfo: %s", unmarshalErr.Error())
		}

		username, usernameErr := getUser(eventInfo.PathWithNamespace)
		if usernameErr != nil {
			return fmt.Sprintf("error while formatting username: %s", usernameErr.Error())
		}

		if readBool("validate_customers") {
			var found bool
			customersURL := os.Getenv("customers_url")
			customers, getErr := getCustomers(customersURL)
			if getErr != nil {
				return getErr.Error()
			}
			found = validCustomer(customers, username)
			if !found {
				auditEvent := sdk.AuditEvent{
					Message: "Customer not found",
					Owner:   username,
					Source:  Source,
				}
				sdk.PostAudit(auditEvent)
				return fmt.Sprintf("Customer: %s not found in CUSTOMERS file via %s", username, customersURL)
			}
		}

		installed, err := InstalledApp(eventInfo.ProjectID, instance, installationTag)
		if err != nil {
			return fmt.Sprintf("Error while trying to connect to GitLab API: %s", err.Error())
		}
		if !installed {
			garbageRequest := []GarbageRequest{}
			garbageRequest = append(garbageRequest,
				GarbageRequest{
					Owner:     username,
					Repo:      eventInfo.Name,
					Functions: []string{},
				})
			err := garbageCollect(garbageRequest)
			if err != nil {
				return fmt.Sprintf("Unexpected error in garbage collect: `%s`\n", err.Error())
			}
			return fmt.Sprintf("Function: `%s` deleted", eventInfo.Name)
		}
	}
	return fmt.Sprintf("Message received with event: %s", eventName.Event)
}

func garbageCollect(garbageRequests []GarbageRequest) error {
	client := http.Client{}

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

		res, err := client.Do(req)
		if err != nil {
			return err
		}
		if res.Body != nil {
			defer res.Body.Close()
		}

		fmt.Printf("Status code returned by garbageCollect for function: `%s` by `%s` - %d\n",
			garbageRequest.Repo, garbageRequest.Owner, res.StatusCode)

		if res.StatusCode != http.StatusAccepted {
			resBody, _ := ioutil.ReadAll(res.Body)
			fmt.Printf("Error in garbageCollect: %s\n", resBody)
		}
	}
	return nil
}

type GarbageRequest struct {
	Functions []string `json:"functions"`
	Repo      string   `json:"repo"`
	Owner     string   `json:"owner"`
}

func forward(req []byte, function string, headers map[string]string) (string, int, error) {
	payloadSecret, err := sdk.ReadSecret("payload-secret")
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	c := http.Client{}

	bodyReader := bytes.NewBuffer(req)
	pushReq, _ := http.NewRequest(http.MethodPost, os.Getenv("gateway_url")+"function/"+function, bodyReader)
	digest := hmac.Sign(req, []byte(payloadSecret))
	pushReq.Header.Add(sdk.CloudSignatureHeader, "sha1="+hex.EncodeToString(digest))

	for k, v := range headers {
		pushReq.Header.Add(k, v)
	}

	res, err := c.Do(pushReq)
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

func InstalledApp(id int, instance, installationTag string) (bool, error) {
	wholeURL := instance + "/api/v4/projects/" + strconv.Itoa(id)

	req, reqErr := http.NewRequest(http.MethodGet, wholeURL, nil)
	if reqErr != nil {
		return false, fmt.Errorf("Error while creating request for GitLab: %s", reqErr.Error())
	}

	client := &http.Client{}
	resp, respErr := client.Do(req)
	if respErr != nil {
		return false, fmt.Errorf("Error while getting response from GitLab: %s", respErr.Error())
	}
	defer resp.Body.Close()

	body, bodyErr := ioutil.ReadAll(resp.Body)
	if bodyErr != nil {
		return false, fmt.Errorf("Error while reading body from GitLab response: %s", bodyErr.Error())
	}

	projectInfo := GitLabProjectTags{}
	unmarshalErr := json.Unmarshal(body, &projectInfo)
	if unmarshalErr != nil {
		return false, fmt.Errorf("Error while un-marshaling projectInfo from body: %s", bodyErr.Error())
	}

	for _, tag := range projectInfo.TagList {
		if strings.EqualFold(tag, installationTag) {
			return true, nil
		}
	}
	return false, nil
}

func readBool(key string) bool {
	if val, exists := os.LookupEnv(key); exists {
		return val == "true" || val == "1"
	}
	return false
}

func getCustomers(customerURL string) ([]string, error) {
	customers := []string{}
	if len(customerURL) == 0 {
		return nil, fmt.Errorf("customerURL was nil")
	}

	httpReq, reqErr := http.NewRequest(http.MethodGet, customerURL, nil)
	if reqErr != nil {
		return nil, fmt.Errorf("error while making the request to `%s` : %s", customerURL, reqErr.Error())
	}

	c := http.Client{}
	res, reqErr := c.Do(httpReq)
	if reqErr != nil {
		return nil, fmt.Errorf("error while requesting customers: %s", reqErr.Error())
	}
	if res.Body != nil {
		defer res.Body.Close()
		pageBody, readErr := ioutil.ReadAll(res.Body)
		if readErr != nil {
			return nil, fmt.Errorf("error while reading response body for customers: %s", readErr)
		}
		customers = strings.Split(string(pageBody), "\n")
		for i, c := range customers {
			customers[i] = strings.ToLower(strings.TrimSuffix(c, "\r"))
		}
	}
	return customers, nil
}

func validCustomer(customers []string, owner string) bool {
	for _, customer := range customers {
		if len(customer) > 0 &&
			strings.EqualFold(customer, owner) {
			return true
		}
	}
	return false
}

func getUser(pathWithNamespace string) (string, error) {
	if !strings.Contains(pathWithNamespace, "/") {
		return "", fmt.Errorf("possible out of range error")
	}
	return pathWithNamespace[:strings.Index(pathWithNamespace, "/")], nil
}

type GitLabProjectEvent struct {
	Name              string `json:"Name"`
	PathWithNamespace string `json:"path_with_namespace"`
	ProjectID         int    `json:"project_id"`
}

type GitLabProjectTags struct {
	TagList []string `json:"tag_list"`
}

type PureEvent struct {
	Event string `json:"event_name"`
}

func checkSupportedEvents(event string) bool {
	for _, supportedEvent := range supportedEvents {
		if supportedEvent == event {
			return true
		}
	}
	return false
}

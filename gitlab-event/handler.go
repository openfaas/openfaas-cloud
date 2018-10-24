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
const Source = "gitlab-event"

// Handle a serverless request
func Handle(req []byte) string {
	eventHeader := os.Getenv("Http_X_Gitlab_Event")
	xGitlabToken := os.Getenv("Http_X_Gitlab_Token")

	if eventHeader != "System Hook" {
		auditEvent := sdk.AuditEvent{
			Message: "required : " + "System Hook",
			Source:  Source,
		}

		sdk.PostAudit(auditEvent)

		return fmt.Sprintf("%s: System Hook required cannot handle: %s", Source, eventHeader)
	}

	eventName := PureEvent{}
	json.Unmarshal(req, &eventName)

	if eventName.Event != "push" &&
		eventName.Event != "project_update" &&
		eventName.Event != "project_destroy" {

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
			return fmt.Sprintf("Unexpected error: %s", secretErr.Error())
		}
		if !tokenMatch(xGitlabToken, tokenSecretKey) {
			return fmt.Sprintf("the request token and the existing tokens mismatch")
		}
	}

	instance := os.Getenv("gitlab_instance")
	if instance == "" {
		return fmt.Sprintf("environmental variable gitlab_instance is missing for gitlab-event")
	}

	apiToken, tokenErr := sdk.ReadSecret("gitlab-api-token")

	if tokenErr != nil {
		return fmt.Sprintf("unable to read GitLab API token: %s", tokenErr.Error())
	}

	switch eventName.Event {
	case "push":
		eventInfo := GitLabPushEvent{}
		unmarshalErr := json.Unmarshal(req, &eventInfo)
		if unmarshalErr != nil {
			return fmt.Sprint("unable to unmarshal request into struct")
		}

		if readBool("validate_customers") {
			var found bool
			customersURL := os.Getenv("customers_url")
			customers, getErr := getCustomers(customersURL)
			if getErr != nil {
				return getErr.Error()
			}
			found = validCustomer(customers, eventInfo.UserUsername)
			if !found {
				auditEvent := sdk.AuditEvent{
					Message: "Customer not found",
					Owner:   eventInfo.UserUsername,
					Source:  Source,
				}
				sdk.PostAudit(auditEvent)
				return fmt.Sprintf("Customer: %s not found in CUSTOMERS file via %s", eventInfo.UserUsername, customersURL)
			}
		}
		installed, err := InstalledApp(eventInfo.GitLabProject.ID, instance, apiToken)
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
			if statusCode == http.StatusOK {
				return fmt.Sprintf("Forwarded to function: `gitlab-push` status: %d response: %s", statusCode, body)
			}
			if err != nil {
				return err.Error()
			}

			return body
		}
		return fmt.Sprintf("This repository is not openfaas-cloud instance")

	case "project_update", "project_destroy":
		eventInfo := GitLabProjectEvent{}
		json.Unmarshal(req, &eventInfo)
		username := getUser(eventInfo.PathWithNamespace)

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

		installed, err := InstalledApp(eventInfo.ProjectID, instance, apiToken)
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

	suffix := os.Getenv("dns_suffix")
	gatewayURL := os.Getenv("gateway_url")
	gatewayURL = sdk.CreateServiceURL(gatewayURL, suffix)

	payloadSecret, err := sdk.ReadSecret("payload-secret")
	if err != nil {
		return err
	}

	for _, garbageRequest := range garbageRequests {

		body, bodyErr := json.Marshal(garbageRequest)
		if bodyErr != nil {
			return fmt.Errorf("error while marshaling garbage-collect request: %s", bodyErr.Error())
		}

		bodyReader := bytes.NewReader(body)
		req, reqErr := http.NewRequest(http.MethodPost, gatewayURL+"async-function/garbage-collect", bodyReader)
		if reqErr != nil {
			return reqErr
		}

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
			resBody, bodyReader := ioutil.ReadAll(res.Body)
			if bodyReader != nil {
				return bodyReader
			}
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

	suffix := os.Getenv("dns_suffix")
	gatewayURL := os.Getenv("gateway_url")
	gatewayURL = sdk.CreateServiceURL(gatewayURL, suffix)

	c := http.Client{}

	bodyReader := bytes.NewBuffer(req)
	pushReq, reqErr := http.NewRequest(http.MethodPost, gatewayURL+"function/"+function, bodyReader)
	if reqErr != nil {
		return "", http.StatusBadRequest, reqErr
	}
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
	body, bodyErr := ioutil.ReadAll(res.Body)
	if bodyErr != nil {
		return "", http.StatusInternalServerError, bodyErr
	}

	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf(string(body))
	}

	return string(body), res.StatusCode, err
}

func InstalledApp(id int, instance, apiToken string) (bool, error) {
	wholeURL := instance + "/api/v4/projects/" + strconv.Itoa(id)

	req, reqErr := http.NewRequest(http.MethodGet, wholeURL, nil)
	if reqErr != nil {
		return false, reqErr
	}
	req.Header.Add("PRIVATE-TOKEN", apiToken)

	client := &http.Client{}
	resp, respErr := client.Do(req)
	if respErr != nil {
		return false, respErr
	}
	defer resp.Body.Close()

	body, bodyErr := ioutil.ReadAll(resp.Body)
	if bodyErr != nil {
		return false, bodyErr
	}
	projectInfo := GitLabProjectTags{}

	unmarshalErr := json.Unmarshal(body, &projectInfo)
	if unmarshalErr != nil {
		return false, unmarshalErr
	}

	for _, tag := range projectInfo.TagList {
		if strings.EqualFold(tag, "openfaas-cloud") {
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

func tokenMatch(gitlabToken string, token string) bool {
	return gitlabToken == token
}

func getCustomers(customerURL string) ([]string, error) {
	customers := []string{}
	if len(customerURL) == 0 {
		return nil, fmt.Errorf("customerURL was nil")
	}
	c := http.Client{}
	httpReq, httpErr := http.NewRequest(http.MethodGet, customerURL, nil)
	if httpErr != nil {
		return customers, httpErr
	}
	res, reqErr := c.Do(httpReq)
	if reqErr != nil {
		return customers, reqErr
	}
	if res.Body != nil {
		defer res.Body.Close()
		pageBody, bodyErr := ioutil.ReadAll(res.Body)
		if bodyErr != nil {
			return customers, bodyErr
		}
		customers = strings.Split(string(pageBody), "\n")
		for i, c := range customers {
			customers[i] = strings.ToLower(c)
			customers[i] = strings.TrimSuffix(c, "\r")
		}
	}
	return customers, nil
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

func getUser(pathWithNamespace string) string {
	return pathWithNamespace[:strings.Index(pathWithNamespace, "/")]
}

type GitLabProjectEvent struct {
	Name              string `json:"Name"`
	PathWithNamespace string `json:"path_with_namespace"`
	ProjectID         int    `json:"project_id"`
}

type GitLabProjectTags struct {
	TagList []string `json:"tag_list"`
}

type GitLabPushEvent struct {
	Event            string           `json:"event_name"`
	Ref              string           `json:"ref"`
	UserUsername     string           `json:"user_username"`
	UserEmail        string           `json:"user_email"`
	GitLabProject    GitLabProject    `json:"project"`
	GitLabRepository GitLabRepository `json:"repository"`
	AfterCommitID    string           `json:"after"`
}
type GitLabProject struct {
	ID                int    `json:"id"`
	Namespace         string `json:"namespace"`
	Name              string `json:"name"`
	PathWithNamespace string `json:"path_with_namespace"` //would be repo full name
	WebURL            string `json:"web_url"`
}
type GitLabRepository struct {
	CloneURL string `json:"git_http_url"`
}

type PureEvent struct {
	Event string `json:"event_name"`
}

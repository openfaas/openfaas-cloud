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
	/*
	 *
	 * Get Headers
	 *
	 */
	eventHeader := os.Getenv("Http_X_Gitlab_Event")
	xGitlabToken := os.Getenv("Http_X_Gitlab_Token")

	/*
	 *
	 * Check bad event from headers above
	 *
	 */

	fmt.Printf("`%q`\n", eventHeader)
	fmt.Printf("`%q`\n", xGitlabToken)

	if eventHeader != "System Hook" {
		fmt.Printf("Good `%q`\n", eventHeader)
		auditEvent := sdk.AuditEvent{
			Message: "required : " + "System Hook",
			Source:  Source,
		}

		sdk.PostAudit(auditEvent)

		return fmt.Sprintf("%s cannot handle event: %s", Source, eventHeader)
	}

	/*
	 *
	 * Check good event from headers above
	 *
	 */

	eventInfo := GitLabEvent{}
	json.Unmarshal(req, &eventInfo)

	if eventInfo.Event != "tag_push" &&
		eventInfo.Event != "push" {

		auditEvent := sdk.AuditEvent{
			Message: "bad event: " + eventInfo.Event,
			Source:  Source,
		}

		sdk.PostAudit(auditEvent)

		return fmt.Sprintf("%s cannot handle event: %s", Source, eventInfo.Event)
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

	instance := getInstance(eventInfo.GitLabProject.WebURL, eventInfo.GitLabProject.PathWithNamespace)

	if InstalledApp(eventInfo.GitLabProject.ID, instance) {
		if readBool("validate_token") {
			tokenSecretKey, secretErr := sdk.ReadSecret("gitlab-webhook-secret")
			if secretErr != nil {
				return secretErr.Error()
			}
			fmt.Printf("\n`%s` == `%s` \n", tokenSecretKey, xGitlabToken)
			if !tokenMatch(xGitlabToken, tokenSecretKey) {
				return fmt.Errorf("tokens don't match, untrusted source\n").Error()
			}
		}

		headers := map[string]string{
			"X-Gitlab-Token": xGitlabToken,
			"X-Gitlab-Event": eventHeader,
			"Content-Type":   "application/json",
		}

		body, statusCode, err := forward(req, "gitlab-push", headers)

		if statusCode == http.StatusOK {
			return fmt.Sprintf("Forwarded to function: %d, %s", statusCode, body)
		}

		if err != nil {
			return err.Error()
		}

		return body
	} else {
		garbageRequests := []GarbageRequest{}
		garbageRequests = append(garbageRequests,
			GarbageRequest{
				Owner:     eventInfo.UserUsername,
				Repo:      eventInfo.GitLabProject.Name,
				Functions: []string{},
			})
		err := garbageCollect(garbageRequests)
		if err != nil {
			return fmt.Sprintf("Unexpected error in garbage collect: `%s`\n", err.Error())
		}
		return fmt.Sprintf("Functions deleted")
	}

	/*
	 *
	 * Check special events from headers above
	 *
	 */
	/*
		if eventHeader == "installation" ||
			eventHeader == "installation_repositories" ||
			eventHeader == "integration_installation" {

			shouldValidate := os.Getenv("validate_hmac")
			if len(shouldValidate) > 0 && (shouldValidate == "1" || shouldValidate == "true") {
				webhookSecretKey, secretErr := sdk.ReadSecret("github-webhook-secret")
				if secretErr != nil {
					return secretErr.Error()
				}

				validateErr := hmac.Validate(req, xHubSignature, webhookSecretKey)
				if validateErr != nil {
					log.Fatal(validateErr)
				}
			}

			event := InstallationRepositoriesEvent{}
			err := json.Unmarshal(req, &event)
			if err != nil {
				return err.Error()
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
	*/

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
			garbageRequest.Owner, garbageRequest.Repo, res.StatusCode)

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

type GitLabEvent struct {
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

func InstalledApp(id int, instance string) bool {
	wholeURL := instance + "/api/v4/projects/" + strconv.Itoa(id)

	req, _ := http.NewRequest(http.MethodGet, wholeURL, nil)

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	projectInfo := GitLabProjectInfo{}

	json.Unmarshal(body, &projectInfo)

	for _, tag := range projectInfo.TagList {
		fmt.Printf("Tag : %s", tag)
		if strings.EqualFold(tag, "install") {
			return true
		}
	}
	return false
}

func getInstance(wholeURL, namespace string) string {
	return strings.TrimSuffix(wholeURL, namespace)
}

type GitLabProjectInfo struct {
	TagList []string `json:"tag_list"`
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

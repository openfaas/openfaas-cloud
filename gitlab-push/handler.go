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

const Source = "gitlab-push"

var audit sdk.Audit

// Handle a serverless request
func Handle(req []byte) string {

	if audit == nil {
		audit = sdk.AuditLogger{}
	}

	event := os.Getenv("Http_X_Gitlab_Event")

	if event != "Push Hook" && event != "System Hook" {
		auditEvent := sdk.AuditEvent{
			Message: "bad event: " + event,
			Source:  Source,
		}
		audit.Post(auditEvent)

		return fmt.Sprintf("%s cannot handle event: %s", Source, event)
	}

	xGitlabToken := os.Getenv("Http_X_Gitlab_Token")

	if readBool("validate_token") {
		tokenSecretKey, secretErr := sdk.ReadSecret("gitlab-webhook-secret")
		if secretErr != nil {
			return secretErr.Error()
		}

		if tokenMatch(xGitlabToken, tokenSecretKey) {
			return fmt.Errorf("gitlab tokens don't match").Error()
		}
	}

	gitlabPushEvent := sdk.GitLabPushEvent{}
	err := json.Unmarshal(req, &gitlabPushEvent)
	if err != nil {
		return err.Error()
	}

	privateRepo := true //:= privateRepo()

	pushEvent := sdk.PushEvent{
		Ref: gitlabPushEvent.Ref,
		Repository: sdk.PushEventRepository{
			Name:     gitlabPushEvent.GitLabProject.Name,
			FullName: gitlabPushEvent.GitLabProject.PathWithNamespace,
			CloneURL: gitlabPushEvent.GitLabRepository.CloneURL,
			Private:  privateRepo,
			Owner: sdk.Owner{
				Login: gitlabPushEvent.GitLabProject.Namespace,
				Email: gitlabPushEvent.UserEmail,
			},
		},
		AfterCommitID: gitlabPushEvent.AfterCommitID,
		Installation: sdk.PushEventInstallation{
			ID: gitlabPushEvent.GitLabProject.ID,
		},
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
			audit.Post(auditEvent)

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

		audit.Post(auditEvent)
		return msg
	}

	serviceValue := fmt.Sprintf("%s-%s", pushEvent.Repository.Owner.Login, pushEvent.Repository.Name)
	eventInfo := sdk.BuildEventFromPushEvent(pushEvent)
	status := sdk.BuildStatus(eventInfo, sdk.EmptyAuthToken)
	status.AddStatus(sdk.StatusPending, fmt.Sprintf("%s stack deploy is in progress", serviceValue), sdk.StackContext)
	reportStatus(status)

	statusCode, postErr := postEvent(pushEvent)
	if postErr != nil {
		status.AddStatus(sdk.StatusFailure, postErr.Error(), sdk.StackContext)
		reportStatus(status)
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

	payloadSecret, err := sdk.ReadSecret("payload-secret")
	if err != nil {
		return http.StatusUnauthorized, err
	}

	body, _ := json.Marshal(pushEvent)

	c := http.Client{}
	bodyReader := bytes.NewBuffer(body)
	httpReq, _ := http.NewRequest(http.MethodPost, gatewayURL+"async-function/git-tar", bodyReader)

	digest := hmac.Sign(body, []byte(payloadSecret))
	httpReq.Header.Add("X-Cloud-Signature", "sha1="+hex.EncodeToString(digest))

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

func reportStatus(status *sdk.Status) {

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

func enableStatusReporting() bool {
	return os.Getenv("report_status") == "true"
}

func tokenMatch(gitlabToken string, token string) bool {
	return gitlabToken == token
}

func privateRepo() bool {
	return os.Getenv("private_repo") == "true"
}

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

	"github.com/alexellis/hmac"
	"github.com/openfaas/openfaas-cloud/sdk"
)

const (
	Source = "gitlab-push"
	SCM    = "gitlab"
)

var audit sdk.Audit

// Handle a serverless request
func Handle(req []byte) string {
	//httpReq.Header.Add(sdk.CloudSignatureHeader, "sha1="+hex.EncodeToString(digest))

	validateErr := validateRequest(req)
	if validateErr != nil {
		log.Fatal(validateErr)
	}

	if audit == nil {
		audit = sdk.AuditLogger{}
	}

	event := os.Getenv("Http_X_Gitlab_Event")

	if event != "System Hook" {
		auditEvent := sdk.AuditEvent{
			Message: "bad event: " + event,
			Source:  Source,
		}
		audit.Post(auditEvent)

		return fmt.Sprintf("%s cannot handle event: %s", Source, event)
	}

	gitlabPushEvent := sdk.GitLabPushEvent{}
	err := json.Unmarshal(req, &gitlabPushEvent)
	if err != nil {
		return err.Error()
	}

	privateRepo := formatPrivateRepo(gitlabPushEvent.GitLabProject.VisibilityLevel)

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
			RepositoryURL: gitlabPushEvent.GitLabProject.WebURL,
		},
		AfterCommitID: gitlabPushEvent.AfterCommitID,
		Installation: sdk.PushEventInstallation{
			ID: gitlabPushEvent.GitLabProject.ID,
		},
	}

	pushEvent.SCM = SCM

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
	statusErr := reportGitLabStatus(status)
	if statusErr != nil {
		log.Printf("error while reporting status: %s", statusErr)
	}

	statusCode, postErr := postEvent(pushEvent)
	if postErr != nil {
		status.AddStatus(sdk.StatusFailure, postErr.Error(), sdk.StackContext)
		statusErr := reportGitLabStatus(status)
		if statusErr != nil {
			log.Printf("error while reporting status: %s", statusErr)
		}
		return postErr.Error()
	}

	auditEvent := sdk.AuditEvent{
		Message: "Git-tar invoked",
		Owner:   pushEvent.Repository.Owner.Login,
		Repo:    pushEvent.Repository.Name,
		Source:  Source,
	}

	sdk.PostAudit(auditEvent)

	return fmt.Sprintf("Push - %v, git-tar status: %d", pushEvent, statusCode)
}

func postEvent(pushEvent sdk.PushEvent) (int, error) {
	suffix := os.Getenv("dns_suffix")
	gatewayURL := os.Getenv("gateway_url")
	gatewayURL = sdk.CreateServiceURL(gatewayURL, suffix)

	payloadSecret, err := sdk.ReadSecret("payload-secret")
	if err != nil {
		return http.StatusUnauthorized, err
	}

	body, bodyErr := json.Marshal(pushEvent)
	if bodyErr != nil {
		return http.StatusBadRequest, fmt.Errorf("error while marshalling event: %s", bodyErr.Error())
	}

	c := http.Client{}
	bodyReader := bytes.NewBuffer(body)
	httpReq, httpErr := http.NewRequest(http.MethodPost, gatewayURL+"async-function/git-tar", bodyReader)
	if httpErr != nil {
		return http.StatusBadRequest, fmt.Errorf("error while making request to git-tar: %s", httpErr.Error())
	}

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

// Needs adding changes to git-lab and merging of gitlab-status function
func reportGitLabStatus(status *sdk.Status) error {

	if !enableStatusReporting() {
		return nil
	}

	payloadSecret, secretErr := sdk.ReadSecret("payload-secret")
	if secretErr != nil {
		return secretErr
	}

	suffix := os.Getenv("dns_suffix")
	gatewayURL := os.Getenv("gateway_url")
	gatewayURL = sdk.CreateServiceURL(gatewayURL, suffix)

	statusBytes, marshalErr := json.Marshal(status)
	if marshalErr != nil {
		return fmt.Errorf("error while marshalling request: %s", marshalErr.Error)
	}

	statusReader := bytes.NewReader(statusBytes)
	req, reqErr := http.NewRequest(http.MethodPost, gatewayURL+"function/gitlab-status", statusReader)
	if reqErr != nil {
		return fmt.Errorf("error while making request to gitlab-status: `%s`", reqErr.Error())
	}

	digest := hmac.Sign(statusBytes, []byte(payloadSecret))
	req.Header.Add(sdk.CloudSignatureHeader, "sha1="+hex.EncodeToString(digest))

	client := http.Client{}

	res, resErr := client.Do(req)
	if resErr != nil {
		return fmt.Errorf("unexpected error while retrieving response: %s", resErr.Error())
	}
	defer res.Body.Close()

	_, bodyErr := ioutil.ReadAll(res.Body)
	if bodyErr != nil {
		return fmt.Errorf("unexpected error while reading response body: %s", bodyErr.Error())
	}

	return nil
}

func enableStatusReporting() bool {
	return os.Getenv("report_status") == "true"
}

func tokenMatch(gitlabToken string, token string) bool {
	return gitlabToken == token
}

func formatPrivateRepo(visibilityLevel int) bool {
	return visibilityLevel != 20
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

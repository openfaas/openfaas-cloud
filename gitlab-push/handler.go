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

const (
	PrivateRepo  = 00
	InternalRepo = 10
	PublicRepo   = 20
	Source       = "gitlab-push"
	SCM          = "gitlab"
)

var audit sdk.Audit

// Handle accepts push event from gitlab-event
// and transforms the payload into PushEvent struct
// which is then sent to git-tar for a function to be built
func Handle(req []byte) string {
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
		return fmt.Sprintf("error while unmarshaling gitlabPushEvent struct: %s", err.Error())
	}

	privateRepo := checkPublicRepo(gitlabPushEvent.GitLabProject.VisibilityLevel)

	pushEvent := sdk.PushEvent{
		SCM: SCM,
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

	eventInfo := sdk.BuildEventFromPushEvent(pushEvent)
	status := sdk.BuildStatus(eventInfo, sdk.EmptyAuthToken)

	branchErr := checkBranch(pushEvent.Ref)
	if branchErr != nil {
		branchErrorMessage := branchErr.Error()
		auditEvent := sdk.AuditEvent{
			Message: branchErrorMessage,
			Owner:   pushEvent.Repository.Owner.Login,
			Repo:    pushEvent.Repository.Name,
			Source:  Source,
		}

		audit.Post(auditEvent)

		status.AddStatus(sdk.StatusFailure, branchErrorMessage, sdk.StackContext)
		reportGitLabStatus(status)
		return branchErrorMessage
	}

	serviceValue := fmt.Sprintf("%s-%s", pushEvent.Repository.Owner.Login, pushEvent.Repository.Name)
	status.AddStatus(sdk.StatusPending, fmt.Sprintf("%s stack deploy is in progress", serviceValue), sdk.StackContext)
	reportGitLabStatus(status)

	statusCode, postErr := postEvent(pushEvent)
	if postErr != nil {
		status.AddStatus(sdk.StatusFailure, postErr.Error(), sdk.StackContext)
		reportGitLabStatus(status)
		return fmt.Sprintf("error while posting event to git-tar: %s", postErr.Error())
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

	bodyReader := bytes.NewBuffer(body)
	httpReq, httpErr := http.NewRequest(http.MethodPost, gatewayURL+"async-function/git-tar", bodyReader)
	if httpErr != nil {
		return http.StatusBadRequest, fmt.Errorf("error while creating request to git-tar: %s", httpErr.Error())
	}
	digest := hmac.Sign(body, []byte(payloadSecret))
	httpReq.Header.Add("X-Cloud-Signature", "sha1="+hex.EncodeToString(digest))

	c := http.Client{}
	res, reqErr := c.Do(httpReq)
	if res.Body != nil {
		defer res.Body.Close()
	}
	if reqErr != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("error while making request to git-tar: %s", reqErr.Error())
	}

	return res.StatusCode, nil
}

func readBool(key string) bool {
	if val, exists := os.LookupEnv(key); exists {
		return val == "true" || val == "1"
	}
	return false
}

func reportGitLabStatus(status *sdk.Status) {
	payloadSecret, secretErr := sdk.ReadSecret("payload-secret")
	if secretErr != nil {
		log.Printf("unexpected error while reading secret: %s", secretErr)
	}
	suffix := os.Getenv("dns_suffix")
	gatewayURL := os.Getenv("gateway_url")
	gatewayURL = sdk.CreateServiceURL(gatewayURL, suffix)
	statusBytes, marshalErr := json.Marshal(status)
	if marshalErr != nil {
		log.Printf("error while marshalling request: %s", marshalErr.Error())
	}

	statusReader := bytes.NewReader(statusBytes)
	req, reqErr := http.NewRequest(http.MethodPost, gatewayURL+"function/gitlab-status", statusReader)
	if reqErr != nil {
		log.Printf("error while making request to gitlab-status: `%s`", reqErr.Error())
	}
	digest := hmac.Sign(statusBytes, []byte(payloadSecret))
	req.Header.Add(sdk.CloudSignatureHeader, "sha1="+hex.EncodeToString(digest))

	client := http.Client{}
	res, resErr := client.Do(req)
	if resErr != nil {
		log.Printf("unexpected error while retrieving response: %s", resErr.Error())
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	if res.StatusCode != http.StatusOK {
		log.Printf("unexpected status code: %d", res.StatusCode)
	}

	_, bodyErr := ioutil.ReadAll(res.Body)
	if bodyErr != nil {
		log.Printf("unexpected error while reading response body: %s", bodyErr.Error())
	}
}

func checkPublicRepo(visibilityLevel int) bool {
	return visibilityLevel != PublicRepo
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

func checkBranch(branchRef string) (branchErr error) {
	buildBranch := getBranch()
	branchFromRef := filterBranchRef(branchRef)
	if buildBranch != branchFromRef {
		msg := fmt.Sprintf("refusing to build target branch: %s, want branch: %s",
			branchFromRef,
			buildBranch)
		branchErr = fmt.Errorf(msg)
	}
	return branchErr
}

func getBranch() string {
	buildBranch := "master"
	if environmentBranch, exists := os.LookupEnv("build_branch"); exists {
		buildBranch = strings.TrimSpace(environmentBranch)
	}
	return buildBranch
}

func filterBranchRef(branchRef string) string {
	stringParts := strings.Split(branchRef, "/")
	branch := "master"
	if len(stringParts) != 0 {
		branch = stringParts[len(stringParts)-1]
	}
	return branch
}

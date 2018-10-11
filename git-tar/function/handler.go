package function

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"encoding/hex"

	"github.com/alexellis/hmac"
	"github.com/openfaas/faas-cli/stack"
	"github.com/openfaas/openfaas-cloud/sdk"
)

// Source of this event for auditing
const Source = "git-tar"

// Handle a serverless request
func Handle(req []byte) []byte {

	start := time.Now()

	shouldValidate := os.Getenv("validate_hmac")

	payloadSecret, secretErr := sdk.ReadSecret("payload-secret")
	if secretErr != nil {
		return []byte(secretErr.Error())
	}

	if len(shouldValidate) > 0 && (shouldValidate == "1" || shouldValidate == "true") {

		cloudHeader := os.Getenv("Http_" + strings.Replace(sdk.CloudSignatureHeader, "-", "_", -1))

		validateErr := hmac.Validate(req, cloudHeader, payloadSecret)
		if validateErr != nil {
			log.Fatal(validateErr)
		}
	}

	pushEvent := sdk.PushEvent{}
	err := json.Unmarshal(req, &pushEvent)
	if err != nil {
		log.Printf("cannot unmarshal git-tar request %s '%s'", err.Error(), string(req))
		os.Exit(-1)
	}

	statusEvent := sdk.BuildEventFromPushEvent(pushEvent)
	status := sdk.BuildStatus(statusEvent, sdk.EmptyAuthToken)

	hasStackFile, getStackFileErr := findStackFile(&pushEvent)

	if getStackFileErr != nil {
		msg := fmt.Sprintf("cannot fetch stack file %s", getStackFileErr.Error())

		status.AddStatus(sdk.StatusFailure, msg, sdk.StackContext)
		reportStatus(status)

		log.Printf(msg)
		os.Exit(-1)
	}

	if !hasStackFile {
		status.AddStatus(sdk.StatusFailure, "unable to find stack.yml", sdk.StackContext)
		reportStatus(status)

		auditEvent := sdk.AuditEvent{
			Message: "no stack.yml file found",
			Owner:   pushEvent.Repository.Owner.Login,
			Repo:    pushEvent.Repository.Name,
			Source:  Source,
		}
		sdk.PostAudit(auditEvent)

		os.Exit(-1)
	}

	clonePath, err := clone(pushEvent)
	if err != nil {
		log.Println("Clone ", err.Error())
		status.AddStatus(sdk.StatusFailure, "clone error : "+err.Error(), sdk.StackContext)
		reportStatus(status)
		os.Exit(-1)
	}

	if _, err := os.Stat(path.Join(clonePath, "template")); err == nil {
		log.Println("Post clone check found a user-generated template folder")
		status.AddStatus(sdk.StatusFailure, "remove custom 'templates' folder", sdk.StackContext)
		reportStatus(status)
		os.Exit(-1)
	}

	stack, err := parseYAML(pushEvent, clonePath)
	if err != nil {
		log.Println("parseYAML ", err.Error())
		status.AddStatus(sdk.StatusFailure, "parseYAML error : "+err.Error(), sdk.StackContext)
		reportStatus(status)
		os.Exit(-1)
	}

	err = fetchTemplates(clonePath)
	if err != nil {
		log.Println("Error fetching templates ", err.Error())
		status.AddStatus(sdk.StatusFailure, "fetchTemplates error : "+err.Error(), sdk.StackContext)
		reportStatus(status)
		os.Exit(-1)
	}

	var shrinkWrapPath string
	shrinkWrapPath, err = shrinkwrap(pushEvent, clonePath)
	if err != nil {
		log.Println("Shrinkwrap ", err.Error())
		status.AddStatus(sdk.StatusFailure, "shrinkwrap error : "+err.Error(), sdk.StackContext)
		reportStatus(status)
		os.Exit(-1)
	}

	var tars []tarEntry
	tars, err = makeTar(pushEvent, shrinkWrapPath, stack)
	if err != nil {
		log.Println("Error creating tar(s): ", err.Error())
		status.AddStatus(sdk.StatusFailure, "tar(s) creation failed, error : "+err.Error(), sdk.StackContext)
		reportStatus(status)
		os.Exit(-1)
	}

	err = importSecrets(pushEvent, stack, clonePath)
	if err != nil {
		log.Printf("Error parsing secrets: %s\n", err.Error())
		status.AddStatus(sdk.StatusFailure, "failed to parse secrets, error : "+err.Error(), sdk.StackContext)
		reportStatus(status)
		os.Exit(-1)
	}

	err = deploy(tars, pushEvent, stack, status, payloadSecret)
	if err != nil {
		status.AddStatus(sdk.StatusFailure, "deploy failed, error : "+err.Error(), sdk.StackContext)
		reportStatus(status)
		log.Printf("deploy error: %s", err)
		os.Exit(-1)
	}

	status.AddStatus(sdk.StatusSuccess, "stack is successfully deployed", sdk.StackContext)
	reportStatus(status)

	err = collect(pushEvent, stack)
	if err != nil {
		log.Printf("collect error: %s", err)
	}

	completed := time.Since(start)

	tarMsg := ""
	for _, tar := range tars {
		tarMsg += fmt.Sprintf("%s @ %s, ", tar.functionName, tar.imageName)
	}

	deploymentMessage := fmt.Sprintf("Deployed: %s. Took %s", strings.TrimRight(tarMsg, ", "), completed.String())

	auditEvent := sdk.AuditEvent{
		Message: deploymentMessage,
		Owner:   pushEvent.Repository.Owner.Login,
		Repo:    pushEvent.Repository.Name,
		Source:  Source,
	}
	sdk.PostAudit(auditEvent)

	return []byte(deploymentMessage)
}

func collect(pushEvent sdk.PushEvent, stack *stack.Services) error {
	var err error

	gatewayURL := os.Getenv("gateway_url")

	garbageReq := GarbageRequest{
		Owner: pushEvent.Repository.Owner.Login,
		Repo:  pushEvent.Repository.Name,
	}

	for k := range stack.Functions {
		garbageReq.Functions = append(garbageReq.Functions, k)
	}

	c := http.Client{
		Timeout: time.Second * 3,
	}

	bytesReq, _ := json.Marshal(garbageReq)
	bufferReader := bytes.NewBuffer(bytesReq)

	payloadSecret, err := getPayloadSecret()

	if err != nil {
		return fmt.Errorf("failed to load payload secret, error %t", err)
	}

	request, _ := http.NewRequest(http.MethodPost, gatewayURL+"function/garbage-collect", bufferReader)

	digest := hmac.Sign(bytesReq, []byte(payloadSecret))

	request.Header.Add(sdk.CloudSignatureHeader, "sha1="+hex.EncodeToString(digest))

	response, err := c.Do(request)

	if err == nil {
		if response.Body != nil {
			defer response.Body.Close()
			bodyBytes, bErr := ioutil.ReadAll(response.Body)
			if bErr != nil {
				log.Fatal(bErr)
			}
			log.Println(string(bodyBytes))
		}
	}

	return err
}

type GarbageRequest struct {
	Functions []string `json:"functions"`
	Repo      string   `json:"repo"`
	Owner     string   `json:"owner"`
}

func enableStatusReporting() bool {
	return os.Getenv("report_status") == "true"
}

func getPayloadSecret() (string, error) {
	payloadSecret, err := sdk.ReadSecret("payload-secret")

	if err != nil {

		return "", fmt.Errorf("failed to load hmac key for status, error %t", err)
	}

	return payloadSecret, nil
}

func reportStatus(status *sdk.Status) {

	if !enableStatusReporting() {
		return
	}

	hmacKey, keyErr := getPayloadSecret()

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

// findStackFile returns true if the repo has a stack.yml file in its git-raw CDN. When
// using a private repo the value will return true always since private repos are not
// available via the CDN. Note: given that the CDN has a 5-minute timeout - this optimization
// may have the undesired effect of preventing a user from deploying within a 5 minute window
// of renaming an incorrect "function.yml" to "stack.yml"
func findStackFile(pushEvent *sdk.PushEvent) (bool, error) {

	// If using a private repo the file will not be available via the git-raw CDN
	if pushEvent.Repository.Private {
		return true, nil
	}

	addr, err := getRawURL(pushEvent.SCM, pushEvent.Repository.RepositoryURL, pushEvent.Repository.Owner.Login, pushEvent.Repository.Name)
	if err != nil {
		return false, err
	}

	req, _ := http.NewRequest(http.MethodHead, addr, nil)
	log.Printf("Stack file request: %s", addr)

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Printf("error finding stack %s", err.Error())

		return false, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}
	log.Printf("Stack file status: %d", res.StatusCode)

	if res.StatusCode == http.StatusOK {
		return true, nil
	}

	return false, nil
}

func getRawURL(scm string, repositoryURL string, repositoryOwnerLogin string, repositoryName string) (string, error) {

	addrPrefix := ""
	rawPath := ""
	switch scm {
	case "github":
		addrPrefix = "https://raw.githubusercontent.com"
	case "gitlab":
		addrPrefix = repositoryURL
		rawPath = "raw/"
	}
	if addrPrefix == "" {
		return "", fmt.Errorf(`failed to find stack.yml file: cannot form proper raw URL.
			Expected pushEvent.SCM to be "github" or "gitlab", but got %s`, scm)
	}

	return fmt.Sprintf("%s/%s/%s/%smaster/stack.yml", addrPrefix, repositoryOwnerLogin, repositoryName, rawPath), nil
}

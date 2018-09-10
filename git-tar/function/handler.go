package function

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"encoding/hex"
	"github.com/alexellis/hmac"
	"github.com/openfaas/faas-cli/stack"
	"github.com/openfaas/openfaas-cloud/sdk"
)

const Source = "git-tar"

// Handle a serverless request
func Handle(req []byte) []byte {

	pushEvent := sdk.PushEvent{}
	err := json.Unmarshal(req, &pushEvent)
	if err != nil {
		log.Printf("cannot unmarshal git-tar request %s '%s'", err.Error(), string(req))
		os.Exit(-1)
	}

	statusEvent := sdk.BuildEventFromPushEvent(pushEvent)
	status := sdk.BuildStatus(statusEvent, sdk.EmptyAuthToken)

	clonePath, err := clone(pushEvent)
	if err != nil {
		log.Println("Clone ", err.Error())
		status.AddStatus(sdk.StatusFailure, "clone error : "+err.Error(), sdk.StackContext)
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

	err = deploy(tars, pushEvent, stack, status)
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

	return []byte(fmt.Sprintf("Deployed tar from: %s", tars))
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

	request.Header.Add("X-Cloud-Signature", "sha1="+hex.EncodeToString(digest))

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

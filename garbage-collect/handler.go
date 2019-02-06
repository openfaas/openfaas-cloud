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
	"time"

	"github.com/alexellis/hmac"
	"github.com/openfaas/openfaas-cloud/sdk"
)

const Source = "garbage-collect"

// Handle function cleans up functions which were removed or renamed
// within the repo for the given user.
func Handle(req []byte) string {
	validateErr := validateRequest(req)

	if validateErr != nil {
		log.Fatal(validateErr)
	}

	garbageReq := GarbageRequest{}
	err := json.Unmarshal(req, &garbageReq)

	if err != nil {
		log.Fatal(err)
	}

	owner := garbageReq.Owner
	if garbageReq.Repo == "*" {
		log.Printf("Removing all functions for %s", owner)
	}

	gatewayURL := os.Getenv("gateway_url")
	deployedFunctions, err := listFunctions(owner, gatewayURL)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Functions owned by %s:\n %s", owner, deployedFunctions)

	deleted := 0
	for _, fn := range deployedFunctions {
		if garbageReq.Repo == "*" ||
			(fn.GetRepo() == garbageReq.Repo && !included(&fn, owner, garbageReq.Functions)) {

			err = deleteFunction(fn.Name, gatewayURL)
			if err != nil {
				auditEvent := sdk.AuditEvent{
					Message: fmt.Sprintf("Unable to delete function: `%s`", fn.Name),
					Source:  Source,
				}
				sdk.PostAudit(auditEvent)

				log.Println(err)
			}
			deleted = deleted + 1
		}
	}

	auditEvent := sdk.AuditEvent{
		Message: fmt.Sprintf("Garbage collection ran for %s/%s - %d functions deleted.", garbageReq.Owner, garbageReq.Repo, deleted),
		Source:  Source,
	}
	sdk.PostAudit(auditEvent)

	return fmt.Sprintf("Garbage collection ran for %s/%s - %d functions deleted.", garbageReq.Owner, garbageReq.Repo, deleted)
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

func formatCloudName(name, owner string) string {
	return owner + "-" + name
}

func included(fn *openFaaSFunction, owner string, functionStack []string) bool {

	for _, name := range functionStack {
		if strings.EqualFold(formatCloudName(name, owner), fn.Name) {
			return true
		}
	}

	return false
}

func deleteFunction(target, gatewayURL string) error {
	var err error
	fmt.Println("Delete ", target)

	c := http.Client{
		Timeout: time.Second * 3,
	}
	delReq := struct {
		FunctionName string
	}{
		FunctionName: target,
	}

	bytesReq, _ := json.Marshal(delReq)
	bufferReader := bytes.NewBuffer(bytesReq)
	httpReq, _ := http.NewRequest(http.MethodDelete, gatewayURL+"system/functions", bufferReader)

	addAuthErr := sdk.AddBasicAuth(httpReq)
	if addAuthErr != nil {
		log.Printf("Basic auth error %s", addAuthErr)
	}

	response, err := c.Do(httpReq)

	if err == nil {
		defer response.Body.Close()
		if response.Body != nil {
			bodyBytes, bErr := ioutil.ReadAll(response.Body)
			if bErr != nil {
				log.Fatal(bErr)
			}
			log.Println(string(bodyBytes))
		}
	}

	return err
}

func listFunctions(owner, gatewayURL string) ([]openFaaSFunction, error) {

	var err error

	c := http.Client{
		Timeout: time.Second * 3,
	}

	request, _ := http.NewRequest(http.MethodGet, gatewayURL+"/function/list-functions?user="+owner, nil)

	response, err := c.Do(request)

	if err == nil {
		defer response.Body.Close()

		if response.Body != nil {
			bodyBytes, bErr := ioutil.ReadAll(response.Body)
			if bErr != nil {
				log.Fatal(bErr)
			}

			functions := []openFaaSFunction{}
			mErr := json.Unmarshal(bodyBytes, &functions)
			if mErr != nil {
				log.Fatal(mErr)
			}

			return functions, nil
		}
	}

	return nil, fmt.Errorf("no functions found for user: %s", owner)
}

type GarbageRequest struct {
	Functions []string `json:"functions"`
	Repo      string   `json:"repo"`
	Owner     string   `json:"owner"`
}

type openFaaSFunction struct {
	Name   string            `json:"name"`
	Image  string            `json:"image"`
	Labels map[string]string `json:"labels"`
}

func (f *openFaaSFunction) GetOwner() string {
	return f.Labels[sdk.FunctionLabelPrefix+"git-owner"]
}

func (f *openFaaSFunction) GetRepo() string {
	return f.Labels[sdk.FunctionLabelPrefix+"git-repo"]
}

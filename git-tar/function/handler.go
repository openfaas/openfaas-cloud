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

	"github.com/openfaas/faas-cli/stack"
	"github.com/openfaas/openfaas-cloud/sdk"
)

// Handle a serverless request
func Handle(req []byte) []byte {

	pushEvent := sdk.PushEvent{}
	err := json.Unmarshal(req, &pushEvent)
	if err != nil {
		log.Println(err.Error())
		os.Exit(-1)
	}

	statusEvent := buildStatusEvent(pushEvent)

	clonePath, err := clone(pushEvent)
	if err != nil {
		log.Println("Clone ", err.Error())
		reportStatus("failure", "Clone"+err.Error(), "BUILD", statusEvent)
		os.Exit(-1)
	}

	stack, err := parseYAML(pushEvent, clonePath)
	if err != nil {
		log.Println("parseYAML ", err.Error())
		reportStatus("failure", "parseYAML"+err.Error(), "BUILD", statusEvent)
		os.Exit(-1)
	}

	var shrinkWrapPath string
	shrinkWrapPath, err = shrinkwrap(pushEvent, clonePath)
	if err != nil {
		log.Println("Shrinkwrap ", err.Error())
		reportStatus("failure", "Shrinkwrap"+err.Error(), "BUILD", statusEvent)
		os.Exit(-1)
	}

	var tars []tarEntry
	tars, err = makeTar(pushEvent, shrinkWrapPath, stack)
	if err != nil {
		log.Println("Error creating tar(s): ", err.Error())
		reportStatus("failure", "Error creating tar(s) : "+err.Error(), "BUILD", statusEvent)
		os.Exit(-1)
	}

	err = deploy(tars, pushEvent, stack)
	if err != nil {
		reportStatus("failure", err.Error(), "BUILD", statusEvent)
		log.Println(err)
	}

	err = collect(pushEvent, stack)
	if err != nil {
		log.Println(err)
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
	request, _ := http.NewRequest(http.MethodPost, gatewayURL+"function/garbage-collect", bufferReader)

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

func reportStatus(status string, desc string, statusContext string, event *sdk.EventInfo) {

	if os.Getenv("report_status") != "true" {
		return
	}

	sdk.ReportStatus(status, desc, statusContext, event)
}

func buildStatusEvent(pushEvent sdk.PushEvent) *sdk.EventInfo {
	info := sdk.EventInfo{}

	info.Service = pushEvent.Repository.Name
	info.Owner = pushEvent.Repository.Owner.Login
	info.Repository = pushEvent.Repository.Name
	info.Sha = pushEvent.AfterCommitID
	info.URL = pushEvent.Repository.CloneURL
	info.InstallationID = pushEvent.Installation.ID

	return &info
}

package function

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/openfaas/openfaas-cloud/sdk"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

// Handle a serverless request
func Handle(req []byte) string {

	garbageReq := GarbageRequest{}
	err := json.Unmarshal(req, &garbageReq)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("garbageReq", garbageReq)
	list, err := listFunctions(garbageReq.Owner)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("list: ", list)

	for _, fn := range list {
		found := false
		for _, deployed := range garbageReq.Functions {
			target := garbageReq.Owner + "-" + deployed
			log.Println(fn, target)

			if fn == target {
				found = true
				break
			}
		}

		if !found {
			err = deleteFunction(garbageReq.Owner + "-" + fn)
			if err != nil {
				log.Println(err)
			}
		}
	}

	auditEvent := sdk.AuditEvent{
		Message: "Garbage Collection Completed",
		Owner:   garbageReq.Owner,
		Repo:    garbageReq.Repo,
		Source:  "garbage-collect",
	}

	sdk.PostAudit(auditEvent)

	return fmt.Sprintf("Garbage collection ran: %s", string(req))
}

func deleteFunction(target string) error {
	var err error
	fmt.Println("Delete ", target)

	gatewayURL := os.Getenv("gateway_url")

	c := http.Client{
		Timeout: time.Second * 3,
	}
	delReq := struct {
		Name string
	}{
		Name: target,
	}

	bytesReq, _ := json.Marshal(delReq)
	bufferReader := bytes.NewBuffer(bytesReq)
	request, _ := http.NewRequest(http.MethodDelete, gatewayURL+"system/functions", bufferReader)

	response, err := c.Do(request)

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

func listFunctions(owner string) ([]string, error) {
	var err error
	var functions []string

	c := http.Client{
		Timeout: time.Second * 3,
	}
	gatewayURL := os.Getenv("gateway_url")
	request, _ := http.NewRequest(http.MethodGet, gatewayURL+"system/functions", nil)

	response, err := c.Do(request)

	if err == nil {
		defer response.Body.Close()
		if response.Body != nil {
			bodyBytes, bErr := ioutil.ReadAll(response.Body)
			if bErr != nil {
				log.Fatal(bErr)
			}

			functionList := []function{}
			mErr := json.Unmarshal(bodyBytes, &functionList)
			if mErr != nil {
				log.Fatal(mErr)
			}
			for _, fn := range functionList {
				functions = append(functions, fn.Name)
			}
		}
	}

	return functions, err
}

type GarbageRequest struct {
	Functions []string `json:"functions"`
	Repo      string   `json:"repo"`
	Owner     string   `json:"owner"`
}

type function struct {
	Name   string            `json:"name"`
	Image  string            `json:"image"`
	Labels map[string]string `json:"labels"`
}

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
	"strconv"
	"strings"
	"time"
)

// Handle a build / deploy request - returns empty string for an error
func Handle(req []byte) string {

	c := &http.Client{}

	builderURL := os.Getenv("builder_url")

	event, eventErr := sdk.BuildEventFromEnv()
	if eventErr != nil {
		log.Panic(eventErr)
	}

	serviceValue := fmt.Sprintf("%s-%s", event.Owner, event.Service)

	log.Printf("%d env-vars for %s", len(event.Environment), serviceValue)

	status := sdk.BuildStatus(event, "")

	reader := bytes.NewBuffer(req)
	res, err := http.Post(builderURL+"build", "application/octet-stream", reader)
	if err != nil {
		fmt.Println(err)
		status.AddStatus(sdk.Failure, err.Error(), sdk.FunctionContext(event.Service))
		reportStatus(status)
		return ""
	}

	defer res.Body.Close()
	buildStatus, _ := ioutil.ReadAll(res.Body)
	imageName := strings.TrimSpace(string(buildStatus))

	repositoryURL := os.Getenv("repository_url")

	if len(repositoryURL) == 0 {
		msg := "repository_url env-var not set"
		fmt.Fprintf(os.Stderr, msg)
		status.AddStatus(sdk.Failure, msg, sdk.FunctionContext(event.Service))
		reportStatus(status)
		os.Exit(1)
	}

	if len(imageName) > 0 {

		// Replace image name for "localhost" for deployment
		imageName = repositoryURL + imageName[strings.Index(imageName, ":"):]

		log.Printf("Deploying %s as %s", imageName, serviceValue)

		defaultMemoryLimit := os.Getenv("default_memory_limit")
		if len(defaultMemoryLimit) == 0 {
			defaultMemoryLimit = "20m"
		}

		deploy := deployment{
			Service: serviceValue,
			Image:   imageName,
			Network: "func_functions",
			Labels: map[string]string{
				"Git-Cloud":      "1",
				"Git-Owner":      event.Owner,
				"Git-Repo":       event.Repository,
				"Git-DeployTime": strconv.FormatInt(time.Now().Unix(), 10), //Unix Epoch string
				"Git-SHA":        event.Sha,
			},
			Limits: Limits{
				Memory: defaultMemoryLimit,
			},
			EnvVars: event.Environment,
		}

		gatewayURL := os.Getenv("gateway_url")

		result, err := deployFunction(deploy, gatewayURL, c)

		if err != nil {
			status.AddStatus(sdk.Failure, err.Error(), sdk.FunctionContext(event.Service))
			reportStatus(status)
			log.Fatal(err.Error())
		}

		log.Println(result)
	}

	status.AddStatus(sdk.Success, fmt.Sprintf("function successfully deployed as: %s", serviceValue), sdk.FunctionContext(event.Service))
	reportStatus(status)

	return fmt.Sprintf("buildStatus %s %s %s", buildStatus, imageName, res.Status)
}

func functionExists(deploy deployment, gatewayURL string, c *http.Client) (bool, error) {

	res, err := http.Get(gatewayURL + "system/functions")

	if err != nil {
	}

	defer res.Body.Close()

	fmt.Println("functionExists status: " + res.Status)
	result, _ := ioutil.ReadAll(res.Body)

	functions := []function{}
	json.Unmarshal(result, &functions)

	for _, function1 := range functions {
		if function1.Name == deploy.Service {
			return true, nil
		}
	}

	return false, err
}

func deployFunction(deploy deployment, gatewayURL string, c *http.Client) (string, error) {
	exists, err := functionExists(deploy, gatewayURL, c)

	bytesOut, _ := json.Marshal(deploy)

	reader := bytes.NewBuffer(bytesOut)

	fmt.Println("Deploying: " + deploy.Image + " as " + deploy.Service)
	var res *http.Response
	var httpReq *http.Request
	var method string
	if exists {
		method = http.MethodPut
	} else {
		method = http.MethodPost
	}

	httpReq, err = http.NewRequest(method, gatewayURL+"system/functions", reader)
	httpReq.Header.Set("Content-Type", "application/json")

	res, err = c.Do(httpReq)

	if err != nil {
		fmt.Println(err)
		return "", err
	}

	defer res.Body.Close()
	fmt.Println("Deploy status: " + res.Status)
	buildStatus, _ := ioutil.ReadAll(res.Body)

	return string(buildStatus), err
}

func reportStatus(status *sdk.Status) {

	if os.Getenv("report_status") != "true" {
		return
	}

	gatewayURL := os.Getenv("gateway_url")

	_, reportErr := status.Report(gatewayURL)
	if reportErr != nil {
		fmt.Printf("failed to report status, error: %s", reportErr.Error())
	}
}

type deployment struct {
	Service string
	Image   string
	Network string
	Labels  map[string]string
	Limits  Limits
	// EnvVars provides overrides for functions.
	EnvVars map[string]string `json:"envVars"`
}

type Limits struct {
	Memory string
}

type function struct {
	Name string
}

package function

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/openfaas/openfaas-cloud/sdk"
)

// Handle a build / deploy request - returns empty string for an error
func Handle(req []byte) string {

	c := &http.Client{}

	builderURL := os.Getenv("builder_url")

	event, eventErr := BuildEventFromEnv()
	if eventErr != nil {
		log.Panic(eventErr)
	}

	auditEvent := sdk.AuditEvent{
		Owner:  event.Owner,
		Repo:   event.Repository,
		Source: "buildshiprun",
	}

	serviceValue := fmt.Sprintf("%s-%s", event.Owner, event.Service)
	log.Printf("%d env-vars for %s", len(event.Environment), serviceValue)

	status := sdk.BuildStatus(event, "")

	reader := bytes.NewBuffer(req)

	r, _ := http.NewRequest(http.MethodPost, builderURL+"build", reader)
	r.Header.Set("Content-Type", "application/octet-stream")

	res, err := c.Do(r)

	if err != nil {
		fmt.Println(err)
		auditEvent.Message = fmt.Sprintf("buildshiprun failure: %s", err.Error())
		sdk.PostAudit(auditEvent)
		status.AddStatus(sdk.Failure, err.Error(), sdk.FunctionContext(event.Service))
		reportStatus(status)
		return auditEvent.Message
	}

	defer res.Body.Close()

	buildStatus, _ := ioutil.ReadAll(res.Body)
	imageName := strings.TrimSpace(string(buildStatus))

	repositoryURL := os.Getenv("repository_url")
	pushRepositoryURL := os.Getenv("push_repository_url")

	if len(repositoryURL) == 0 {
		msg := "repository_url env-var not set"
		fmt.Fprintf(os.Stderr, msg)
		status.AddStatus(sdk.Failure, msg, sdk.FunctionContext(event.Service))
		reportStatus(status)
		os.Exit(1)
	}

	if len(pushRepositoryURL) == 0 {
		fmt.Fprintf(os.Stderr, "push_repository_url env-var not set")
		os.Exit(1)
	}

	log.Printf("buildshiprun: image '%s'\n", imageName)

	if strings.Contains(imageName, "exit status") == true {
		msg := "Unable to build image, check builder logs"
		status.AddStatus(sdk.Failure, msg, sdk.FunctionContext(event.Service))
		reportStatus(status)
		log.Fatal(msg)
		auditEvent.Message = fmt.Sprintf("buildshiprun failure: %s", msg)
		sdk.PostAudit(auditEvent)
		return msg
	}

	if len(imageName) > 0 {

		// Replace image name for "localhost" for deployment
		imageName = getImageName(repositoryURL, pushRepositoryURL, imageName)

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
				"faas_function":  serviceValue,
				"app":            serviceValue,
			},
			Limits: Limits{
				Memory: defaultMemoryLimit,
			},
			EnvVars: event.Environment,
			Secrets: event.Secrets,
		}

		gatewayURL := os.Getenv("gateway_url")

		result, err := deployFunction(deploy, gatewayURL, c)

		if err != nil {
			status.AddStatus(sdk.Failure, err.Error(), sdk.FunctionContext(event.Service))
			reportStatus(status)
			log.Fatal(err.Error())
			auditEvent.Message = fmt.Sprintf("buildshiprun failure: %s", err.Error())
		} else {
			auditEvent.Message = fmt.Sprintf("buildshiprun succeeded: deployed %s", imageName)
		}

		log.Println(result)
	}

	sdk.PostAudit(auditEvent)
	status.AddStatus(sdk.Success, fmt.Sprintf("function successfully deployed as: %s", serviceValue), sdk.FunctionContext(event.Service))
	reportStatus(status)
	return fmt.Sprintf("buildStatus %s %s %s", buildStatus, imageName, res.Status)
}

func functionExists(deploy deployment, gatewayURL string, c *http.Client) (bool, error) {

	r, _ := http.NewRequest(http.MethodGet, gatewayURL+"system/functions", nil)

	addAuthErr := sdk.AddBasicAuth(r)
	if addAuthErr != nil {
		log.Printf("Basic auth error %s", addAuthErr)
	}

	res, err := c.Do(r)

	if err != nil {
		fmt.Println(err)
		return false, err
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

	addAuthErr := sdk.AddBasicAuth(httpReq)
	if addAuthErr != nil {
		log.Printf("Basic auth error %s", addAuthErr)
	}

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

func enableStatusReporting() bool {
	return os.Getenv("report_status") == "true"
}

func reportStatus(status *sdk.Status) {
	if !enableStatusReporting() {
		return
	}

	gatewayURL := os.Getenv("gateway_url")

	_, reportErr := status.Report(gatewayURL)
	if reportErr != nil {
		log.Printf("failed to report status, error: %s", reportErr.Error())
	}
}

func getImageName(repositoryURL, pushRepositoryURL, imageName string) string {

	return strings.Replace(imageName, pushRepositoryURL, repositoryURL, 1)

	// return repositoryURL + imageName[strings.Index(imageName, "/"):]
}

// function to build Event from Environment Variable
func BuildEventFromEnv() (*sdk.Event, error) {
	var err error
	info := sdk.Event{}

	info.Service = os.Getenv("Http_Service")
	info.Owner = os.Getenv("Http_Owner")
	info.Repository = os.Getenv("Http_Repo")
	info.Sha = os.Getenv("Http_Sha")
	info.URL = os.Getenv("Http_Url")
	info.InstallationID, err = strconv.Atoi(os.Getenv("Http_Installation_id"))
	info.Environment = GetEnv(info.Service)
	info.Secrets = GetSecret(info.Owner, info.Service)

	return &info, err
}

func GetEnv(service string) map[string]string {
	envVars := make(map[string]string)
	envStr := os.Getenv("Http_Env")
	if len(envStr) > 0 {
		envErr := json.Unmarshal([]byte(envStr), &envVars)
		if envErr != nil {
			log.Printf("error un-marshaling env-vars for function %s, %s", service, envErr)
		}
	}
	return envVars
}

func GetSecret(owner, service string) []string {
	secretVars := []string{}
	secretsStr := os.Getenv("Http_Secrets")
	if len(secretsStr) > 0 {
		secretErr := json.Unmarshal([]byte(secretsStr), &secretVars)
		if secretErr != nil {
			log.Println("error un-marshaling env-vars for function %s, %s", service, secretErr)
		}
	}
	for i := 0; i < len(secretVars); i++ {
		secretVars[i] = owner + "-" + secretVars[i]
	}
	return secretVars
}

type eventInfo struct {
	service        string
	owner          string
	repository     string
	image          string
	sha            string
	url            string
	installationID int
	environment    map[string]string
	secrets        []string
}

type deployment struct {
	Service string
	Image   string
	Network string
	Labels  map[string]string
	Limits  Limits
	// EnvVars provides overrides for functions.
	EnvVars map[string]string `json:"envVars"`
	Secrets []string
}

type Limits struct {
	Memory string
}

type function struct {
	Name string
}

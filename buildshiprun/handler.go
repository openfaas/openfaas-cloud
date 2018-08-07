package function

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/openfaas/openfaas-cloud/sdk"
)

var (
	imageValidator = regexp.MustCompile("(?:[a-zA-Z0-9./]*(?:[._-][a-z0-9]?)*(?::[0-9]+)?[a-zA-Z0-9./]+(?:[._-][a-z0-9]+)*/)*[a-zA-Z0-9]+(?:[._-][a-z0-9]+)+(?::[a-zA-Z0-9._-]+)?")
)

// Handle a build / deploy request - returns empty string for an error
func Handle(req []byte) string {

	c := &http.Client{}

	builderURL := os.Getenv("builder_url")

	event, eventErr := getEvent()
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

	status := sdk.BuildStatus(event, sdk.EmptyAuthToken)

	reader := bytes.NewBuffer(req)

	r, _ := http.NewRequest(http.MethodPost, builderURL+"build", reader)
	r.Header.Set("Content-Type", "application/octet-stream")

	res, err := c.Do(r)

	log.Printf("Image build status: %d\n", res.StatusCode)

	if err != nil {
		fmt.Println(err)
		auditEvent.Message = fmt.Sprintf("buildshiprun failure: %s", err.Error())
		sdk.PostAudit(auditEvent)
		status.AddStatus(sdk.StatusFailure, err.Error(), sdk.BuildFunctionContext(event.Service))
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
		status.AddStatus(sdk.StatusFailure, msg, sdk.BuildFunctionContext(event.Service))
		reportStatus(status)
		os.Exit(1)
	}

	if len(pushRepositoryURL) == 0 {
		fmt.Fprintf(os.Stderr, "push_repository_url env-var not set")
		os.Exit(1)
	}

	log.Printf("buildshiprun: image '%s'\n", imageName)

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusAccepted {
		msg := "Unable to build image, check builder logs"
		status.AddStatus(sdk.StatusFailure, msg, sdk.BuildFunctionContext(event.Service))
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

		readOnlyRootFS := getReadOnlyRootFS()

		registryAuth := getRegistryAuthSecret()

		deploy := deployment{
			Service: serviceValue,
			Image:   imageName,
			Network: "func_functions",
			Labels: map[string]string{
				"Git-Cloud":      "1",
				"Git-Owner":      event.Owner,
				"Git-Repo":       event.Repository,
				"Git-DeployTime": strconv.FormatInt(time.Now().Unix(), 10), //Unix Epoch string
				"Git-SHA":        event.SHA,
				"faas_function":  serviceValue,
				"app":            serviceValue,
			},
			Limits: Limits{
				Memory: defaultMemoryLimit,
			},
			EnvVars:                event.Environment,
			Secrets:                event.Secrets,
			ReadOnlyRootFilesystem: readOnlyRootFS,
		}

		gatewayURL := os.Getenv("gateway_url")

		if len(registryAuth) > 0 {
			deploy.RegistryAuth = registryAuth
		}

		result, err := deployFunction(deploy, gatewayURL, c)

		if err != nil {
			status.AddStatus(sdk.StatusFailure, err.Error(), sdk.BuildFunctionContext(event.Service))
			reportStatus(status)
			log.Fatal(err.Error())
			auditEvent.Message = fmt.Sprintf("buildshiprun failure: %s", err.Error())
		} else {
			auditEvent.Message = fmt.Sprintf("buildshiprun succeeded: deployed %s", imageName)
		}

		log.Println(result)
	}

	sdk.PostAudit(auditEvent)
	status.AddStatus(sdk.StatusSuccess, fmt.Sprintf("function successfully deployed as: %s", serviceValue), sdk.BuildFunctionContext(event.Service))
	reportStatus(status)
	return fmt.Sprintf("buildStatus %s %s %s", buildStatus, imageName, res.Status)
}

// readOnlyRootFS defaults to true, override with env-var of readonly_root_filesystem=false
func getReadOnlyRootFS() bool {
	readOnly := true
	if val, exists := os.LookupEnv("readonly_root_filesystem"); exists {
		if val == "0" || val == "false" {
			readOnly = false
		}
	}

	return readOnly
}

func getEvent() (*sdk.Event, error) {
	var err error
	info := sdk.Event{}

	info.Service = os.Getenv("Http_Service")
	info.Owner = os.Getenv("Http_Owner")
	info.Repository = os.Getenv("Http_Repo")
	info.SHA = os.Getenv("Http_Sha")
	info.URL = os.Getenv("Http_Url")
	info.Image = os.Getenv("Http_Image")

	if len(os.Getenv("Http_Installation_id")) > 0 {
		info.InstallationID, err = strconv.Atoi(os.Getenv("Http_Installation_id"))
	}

	httpEnv := os.Getenv("Http_Env")
	envVars := make(map[string]string)

	if len(httpEnv) > 0 {
		envErr := json.Unmarshal([]byte(httpEnv), &envVars)

		if envErr == nil {
			info.Environment = envVars
		} else {
			log.Printf("Error un-marshaling env-vars for function %s, %s", info.Service, envErr)
			info.Environment = make(map[string]string)
		}
	}

	secretVars := []string{}
	secretsStr := os.Getenv("Http_Secrets")

	if len(secretsStr) > 0 {
		secretErr := json.Unmarshal([]byte(secretsStr), &secretVars)

		if secretErr != nil {
			log.Println(secretErr)
		}

	}

	info.Secrets = secretVars

	for i := 0; i < len(info.Secrets); i++ {
		info.Secrets[i] = info.Owner + "-" + info.Secrets[i]
	}

	log.Printf("%d env-vars for %s", len(info.Environment), info.Service)

	return &info, err
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
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return "", fmt.Errorf("http status code %d", res.StatusCode)
	}
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

func validImage(image string) bool {
	if len(image) <= 0 {
		return false
	}
	match := imageValidator.FindString(image)
	// image should be the whole string
	if len(match) == len(image) {
		return true
	}
	return false
}

type deployment struct {
	Service string
	Image   string
	Network string
	Labels  map[string]string
	Limits  Limits
	// EnvVars provides overrides for functions.
	EnvVars                map[string]string `json:"envVars"`
	Secrets                []string
	ReadOnlyRootFilesystem bool   `json:"readOnlyRootFilesystem"`
	RegistryAuth           string `json:"registryAuth"`
}

type Limits struct {
	Memory string
}

type function struct {
	Name string
}

func getRegistryAuthSecret() string {
	path := "/var/openfaas/secrets/registry-auth"
	if _, err := os.Stat(path); err == nil {
		res, readErr := ioutil.ReadFile(path)
		if readErr != nil {
			log.Printf("Tried to read secret %s, but got error: %s\n", path, readErr)
		}
		return strings.TrimSpace(string(res))
	}
	return ""
}

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
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/alexellis/hmac"
	"github.com/openfaas/openfaas-cloud/sdk"
)

var (
	imageValidator = regexp.MustCompile("(?:[a-zA-Z0-9./]*(?:[._-][a-z0-9]?)*(?::[0-9]+)?[a-zA-Z0-9./]+(?:[._-][a-z0-9]+)*/)*[a-zA-Z0-9]+(?:[._-][a-z0-9]+)+(?::[a-zA-Z0-9._-]+)?")
)

func validateRequest(req *[]byte) (err error) {
	payloadSecret, err := sdk.ReadSecret("payload-secret")

	if err != nil {
		return fmt.Errorf("couldn't get payload-secret: %t", err)
	}

	xCloudSignature := os.Getenv("Http_X_Cloud_Signature")

	err = hmac.Validate(*req, xCloudSignature, payloadSecret)

	if err != nil {
		return err
	}

	return nil
}

// Handle submits the tar to the of-builder then configures an OpenFaaS deployment based upon stack.yml found in the Git repo.
// Finally starts a rolling deployment of the function.
func Handle(req []byte) string {

	hmacErr := validateRequest(&req)
	if hmacErr != nil {
		return fmt.Sprintf("invalid HMAC digest for tar: %s", hmacErr.Error())
	}

	c := &http.Client{}

	builderURL := os.Getenv("builder_url")
	gatewayURL := os.Getenv("gateway_url")

	payloadSecret, keyErr := sdk.ReadSecret("payload-secret")
	if keyErr != nil {
		err := fmt.Errorf("failed to load hmac key, error %s", keyErr.Error())
		log.Printf(err.Error())
		return err.Error()
	}

	event, eventErr := getEvent()
	if eventErr != nil {
		log.Panic(eventErr)
	}

	auditEvent := sdk.AuditEvent{
		Owner:  event.Owner,
		Repo:   event.Repository,
		Source: "buildshiprun",
	}

	serviceValue := sdk.FormatServiceName(event.Owner, event.Service)
	log.Printf("%d env-vars for %s", len(event.Environment), serviceValue)

	status := sdk.BuildStatus(event, sdk.EmptyAuthToken)

	reader := bytes.NewBuffer(req)

	xCloudSignature := os.Getenv("Http_X_Cloud_Signature")

	r, _ := http.NewRequest(http.MethodPost, builderURL+"build", reader)

	r.Header.Set(sdk.CloudSignatureHeader, xCloudSignature)
	r.Header.Set("Content-Type", "application/octet-stream")

	res, err := c.Do(r)

	if err != nil {
		log.Printf("of-builder error: %s\n", err)

		auditEvent.Message = fmt.Sprintf("buildshiprun failure: %s", err.Error())
		sdk.PostAudit(auditEvent)

		status.AddStatus(sdk.StatusFailure, err.Error(), sdk.BuildFunctionContext(event.Service))
		reportStatus(status)

		return auditEvent.Message
	}

	log.Printf("Image build status: %d\n", res.StatusCode)

	defer res.Body.Close()

	buildBytes, _ := ioutil.ReadAll(res.Body)

	result := sdk.BuildResult{}
	unmarshalErr := json.Unmarshal(buildBytes, &result)

	if unmarshalErr != nil {
		log.Printf("BuildResult unmarshalErr %s\n", err)

		auditEvent.Message = fmt.Sprintf("buildshiprun failure reading response: %s, response: %s", unmarshalErr.Error(), string(buildBytes))
		sdk.PostAudit(auditEvent)

		status.AddStatus(sdk.StatusFailure, err.Error(), sdk.BuildFunctionContext(event.Service))
		reportStatus(status)
		return auditEvent.Message
	}

	imageName := strings.ToLower(result.ImageName)

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

	logStatus, logErr := createPipelineLog(result, event, gatewayURL, c, payloadSecret)
	if logErr != nil {
		log.Printf("pipeline-log: error: %s", err.Error())
	} else {
		log.Printf("pipeline-log: status: %d", logStatus)
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusAccepted {
		msg := "Unable to build image, check builder logs"
		status.AddStatus(sdk.StatusFailure, msg, sdk.BuildFunctionContext(event.Service))
		reportStatus(status)

		auditEvent.Message = fmt.Sprintf("buildshiprun failure: %s", msg)
		sdk.PostAudit(auditEvent)

		log.Printf("of-builder result: %s, logs: %s\n", result.Status, strings.Join(result.Log, "\n"))

		log.Fatal(msg)
		return msg
	}

	if len(imageName) > 0 {

		// Replace image name for "localhost" for deployment
		imageName = getImageName(repositoryURL, pushRepositoryURL, imageName)

		log.Printf("Deploying %s as %s", imageName, serviceValue)

		defaultMemoryLimit := getMemoryLimit()

		scalingMinLimit := getConfig("scaling_min_limit", "1")
		scalingMaxLimit := getConfig("scaling_max_limit", "4")

		readOnlyRootFS := getReadOnlyRootFS()

		registryAuth := getRegistryAuthSecret()

		private := 0
		if event.Private {
			private = 1
		}

		scm := "github" // TODO: read other scm values such as GitLab

		deploy := deployment{
			Service: serviceValue,
			Image:   imageName,
			Network: "func_functions",
			Labels: map[string]string{
				sdk.FunctionLabelPrefix + "git-cloud":      "1",
				sdk.FunctionLabelPrefix + "git-owner":      event.Owner,
				sdk.FunctionLabelPrefix + "git-repo":       event.Repository,
				sdk.FunctionLabelPrefix + "git-deploytime": strconv.FormatInt(time.Now().Unix(), 10), //Unix Epoch string
				sdk.FunctionLabelPrefix + "git-sha":        event.SHA,
				sdk.FunctionLabelPrefix + "git-private":    fmt.Sprintf("%d", private),
				sdk.FunctionLabelPrefix + "git-scm":        scm,
				"faas_function":                            serviceValue,
				"app":                                      serviceValue,
				"com.openfaas.scale.min": scalingMinLimit,
				"com.openfaas.scale.max": scalingMaxLimit,
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

		deployResult, err := deployFunction(deploy, gatewayURL, c)

		if err != nil {
			status.AddStatus(sdk.StatusFailure, err.Error(), sdk.BuildFunctionContext(event.Service))
			reportStatus(status)
			log.Fatal(err.Error())
			auditEvent.Message = fmt.Sprintf("buildshiprun failure: %s", err.Error())
		} else {
			auditEvent.Message = fmt.Sprintf("buildshiprun succeeded: deployed %s", imageName)
		}

		log.Println(deployResult)
	}

	sdk.PostAudit(auditEvent)
	status.AddStatus(sdk.StatusSuccess, fmt.Sprintf("deployed: %s", serviceValue), sdk.BuildFunctionContext(event.Service))
	reportStatus(status)
	return fmt.Sprintf("buildStatus %s %s", imageName, res.Status)
}

func getConfig(key string, defaultValue string) string {

	res := os.Getenv(key)
	if len(res) == 0 {
		res = defaultValue
	}
	return res
}

// createPipelineLog sends a log to pipeline-log and will
// fail silently if unavailable.
func createPipelineLog(result sdk.BuildResult, event *sdk.Event, gatewayURL string, c *http.Client, payloadSecret string) (int, error) {

	p := sdk.PipelineLog{
		CommitSHA: event.SHA,
		Function:  event.Service,
		RepoPath:  event.Owner + "/" + event.Repository,
		Data:      strings.Join(result.Log, "\n"),
	}

	bytesOut, _ := json.Marshal(&p)

	reader := bytes.NewReader(bytesOut)

	req, _ := http.NewRequest(http.MethodPost, gatewayURL+"function/pipeline-log", reader)

	digest := hmac.Sign(bytesOut, []byte(payloadSecret))
	req.Header.Add(sdk.CloudSignatureHeader, "sha1="+hex.EncodeToString(digest))

	res, err := c.Do(req)
	if err != nil {
		return http.StatusInternalServerError, err

	}

	if req.Body != nil {
		defer req.Body.Close()
	}

	return res.StatusCode, nil
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

	payloadSecret, keyErr := sdk.ReadSecret("payload-secret")
	if keyErr != nil {
		log.Printf("failed to load hmac key for status, error " + keyErr.Error())
		return
	}

	_, reportErr := status.Report(gatewayURL, payloadSecret)
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
	path := "/var/openfaas/secrets/swarm-pull-secret"
	if _, err := os.Stat(path); err == nil {
		res, readErr := ioutil.ReadFile(path)
		if readErr != nil {
			log.Printf("Tried to read secret %s, but got error: %s\n", path, readErr)
		}
		return strings.TrimSpace(string(res))
	}
	return ""
}

func getMemoryLimit() string {
	const swarmSuffix = "m"
	const kubernetesSuffix = "Mi"

	suffix := swarmSuffix

	kubernetesPort := "KUBERNETES_SERVICE_PORT"
	memoryLimit := os.Getenv("function_memory_limit_mb")

	if _, exists := os.LookupEnv(kubernetesPort); exists {
		suffix = kubernetesSuffix
	}

	const defaultMemoryLimit = "128"

	unit := defaultMemoryLimit
	if len(memoryLimit) > 0 {
		unit = memoryLimit
	}

	return fmt.Sprintf("%s%s", unit, suffix)
}

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

const (
	// GitHub SCM
	GitHub = "github"
	// GitLab SCM
	GitLab = "gitlab"
)

const scaleToZeroDefault = true
const zeroScaleLabel = "com.openfaas.scale.zero"

var (
	imageValidator = regexp.MustCompile("(?:[a-zA-Z0-9./]*(?:[._-][a-z0-9]?)*(?::[0-9]+)?[a-zA-Z0-9./]+(?:[._-][a-z0-9]+)*/)*[a-zA-Z0-9]+(?:[._-][a-z0-9]+)+(?::[a-zA-Z0-9._-]+)?")
)

type FunctionResources struct {
	Memory string `json:"memory,omitempty"`
	CPU    string `json:"cpu,omitempty"`
}

type CPULimits struct {
	Limit     string
	Requests  string
	Available bool
}

// Handle submits the tar to the of-builder then configures an OpenFaaS
// deployment based upon stack.yml found in the Git repo. Finally starts
// a rolling deployment of the function.
func Handle(req []byte) string {

	hmacErr := validateRequest(&req)
	if hmacErr != nil {
		return fmt.Sprintf("invalid HMAC digest for tar: %s", hmacErr.Error())
	}

	builderURL := os.Getenv("builder_url")
	gatewayURL := os.Getenv("gateway_url")

	payloadSecret, keyErr := sdk.ReadSecret("payload-secret")
	if keyErr != nil {
		err := fmt.Errorf("failed to load hmac key, error %s", keyErr.Error())
		log.Printf(err.Error())
		return err.Error()
	}

	event, eventErr := getEventFromEnv()
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

	res, err := http.DefaultClient.Do(r)

	if err != nil {
		log.Printf("of-builder error: %s\n", err)

		auditEvent.Message = fmt.Sprintf("buildshiprun failure: %s", err.Error())
		sdk.PostAudit(auditEvent)

		status.AddStatus(sdk.StatusFailure, err.Error(), sdk.BuildFunctionContext(event.Service))
		statusErr := reportStatus(status, event.SCM)
		if statusErr != nil {
			log.Printf(statusErr.Error())
		}

		return auditEvent.Message
	}

	log.Printf("Image build status: %d\n", res.StatusCode)

	defer res.Body.Close()

	buildBytes, _ := ioutil.ReadAll(res.Body)

	result := sdk.BuildResult{}
	unmarshalErr := json.Unmarshal(buildBytes, &result)

	if unmarshalErr != nil {
		log.Printf("BuildResult unmarshalErr %s\n", unmarshalErr)

		auditEvent.Message = fmt.Sprintf("buildshiprun failure reading response: %s, response: %s", unmarshalErr.Error(), string(buildBytes))
		sdk.PostAudit(auditEvent)

		status.AddStatus(sdk.StatusFailure, unmarshalErr.Error(), sdk.BuildFunctionContext(event.Service))
		statusErr := reportStatus(status, event.SCM)
		if statusErr != nil {
			log.Printf(statusErr.Error())
		}
		return auditEvent.Message
	}

	imageName := strings.ToLower(result.ImageName)

	repositoryURL := os.Getenv("repository_url")
	pushRepositoryURL := os.Getenv("push_repository_url")

	if len(repositoryURL) == 0 {
		msg := "repository_url env-var not set"
		fmt.Fprintf(os.Stderr, msg)
		status.AddStatus(sdk.StatusFailure, msg, sdk.BuildFunctionContext(event.Service))
		statusErr := reportStatus(status, event.SCM)
		if statusErr != nil {
			log.Printf(statusErr.Error())
		}
		os.Exit(1)
	}

	if len(pushRepositoryURL) == 0 {
		fmt.Fprintf(os.Stderr, "push_repository_url env-var not set")
		os.Exit(1)
	}

	log.Printf("buildshiprun: image '%s'\n", imageName)

	logStatus, logErr := createPipelineLog(result, event, gatewayURL, payloadSecret)
	if logErr != nil {
		log.Printf("pipeline-log: error: %s", err.Error())
	} else {
		log.Printf("pipeline-log: status: %d", logStatus)
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusAccepted {
		msg := "Unable to build image, check builder logs"
		status.AddStatus(sdk.StatusFailure, msg, sdk.BuildFunctionContext(event.Service))
		statusErr := reportStatus(status, event.SCM)
		if statusErr != nil {
			log.Printf(statusErr.Error())
		}

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

		scalingFactor := getConfig("scaling_factor", "20")

		readOnlyRootFS := getReadOnlyRootFS()

		registryAuth := getRegistryAuthSecret()

		private := 0
		if event.Private {
			private = 1
		}

		scaleToZero := scaleToZeroDefault

		if val, ok := event.Labels[zeroScaleLabel]; ok && len(val) > 0 {
			boolVal, err := strconv.ParseBool(val)
			if err != nil {
				log.Printf("error parsing label %s : %s", zeroScaleLabel, err.Error())
			} else {
				scaleToZero = boolVal
			}
		}

		annotationWhitelist := []string{
			"topic",
			"schedule",
		}
		userAnnotations := buildAnnotations(annotationWhitelist, event.Annotations)
		userAnnotations[sdk.FunctionLabelPrefix+"git-repo-url"] = event.RepoURL

		deploy := deployment{
			Service: serviceValue,
			Image:   imageName,
			Network: "func_functions",
			Labels: map[string]string{
				sdk.FunctionLabelPrefix + "git-cloud":      "1",
				sdk.FunctionLabelPrefix + "git-owner":      event.Owner,
				sdk.FunctionLabelPrefix + "git-owner-id":   fmt.Sprintf("%d", event.OwnerID),
				sdk.FunctionLabelPrefix + "git-repo":       event.Repository,
				sdk.FunctionLabelPrefix + "git-deploytime": strconv.FormatInt(time.Now().Unix(), 10), //Unix Epoch string
				sdk.FunctionLabelPrefix + "git-sha":        event.SHA,
				sdk.FunctionLabelPrefix + "git-private":    fmt.Sprintf("%d", private),
				sdk.FunctionLabelPrefix + "git-scm":        event.SCM,
				sdk.FunctionLabelPrefix + "git-branch":     buildBranch(),
				"faas_function":                            serviceValue,
				"app":                                      serviceValue,
				"com.openfaas.scale.min":    scalingMinLimit,
				"com.openfaas.scale.max":    scalingMaxLimit,
				"com.openfaas.scale.factor": scalingFactor,
				zeroScaleLabel:              strconv.FormatBool(scaleToZero),
			},
			Annotations:            userAnnotations,
			Requests:               &FunctionResources{},
			Limits:                 &FunctionResources{},
			EnvVars:                event.Environment,
			Secrets:                event.Secrets,
			ReadOnlyRootFilesystem: readOnlyRootFS,
		}

		deploy.Limits.Memory = defaultMemoryLimit

		cpuLimit := getCPULimit()
		if cpuLimit.Available {

			if len(cpuLimit.Limit) > 0 {
				deploy.Limits.CPU = cpuLimit.Limit
			}

			if len(cpuLimit.Requests) > 0 {
				deploy.Requests.CPU = cpuLimit.Requests
			}
		}

		gatewayURL := os.Getenv("gateway_url")

		if len(registryAuth) > 0 {
			deploy.RegistryAuth = registryAuth
		}

		deployResult, err := deployFunction(deploy, gatewayURL)

		log.Println(deployResult)

		if err != nil {
			status.AddStatus(sdk.StatusFailure, err.Error(), sdk.BuildFunctionContext(event.Service))
			statusErr := reportStatus(status, event.SCM)
			if statusErr != nil {
				log.Printf(statusErr.Error())
			}
			log.Fatal(err.Error())
			auditEvent.Message = fmt.Sprintf("buildshiprun failure: %s", err.Error())
			sdk.PostAudit(auditEvent)
			log.Fatalf("buildshiprun failure: %s", err.Error())
		} else {
			auditEvent.Message = fmt.Sprintf("buildshiprun succeeded: deployed %s", imageName)
			sdk.PostAudit(auditEvent)
		}

	}

	status.AddStatus(sdk.StatusSuccess, fmt.Sprintf("deployed: %s", serviceValue), sdk.BuildFunctionContext(event.Service))
	statusErr := reportStatus(status, event.SCM)
	if statusErr != nil {
		log.Printf(statusErr.Error())
	}
	return fmt.Sprintf("buildStatus %s %s", imageName, res.Status)
}

func buildAnnotations(whitelist []string, userValues map[string]string) map[string]string {
	annotations := map[string]string{}
	for k, v := range userValues {
		for _, allowable := range whitelist {
			if allowable == k {
				annotations[k] = v
			}
		}
	}

	return annotations
}

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

func getConfig(key string, defaultValue string) string {

	res := os.Getenv(key)
	if len(res) == 0 {
		res = defaultValue
	}
	return res
}

// createPipelineLog sends a log to pipeline-log and will
// fail silently if unavailable.
func createPipelineLog(result sdk.BuildResult, event *sdk.Event, gatewayURL string, payloadSecret string) (int, error) {

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

	res, err := http.DefaultClient.Do(req)
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

func getEventFromEnv() (*sdk.Event, error) {
	var err error
	info := sdk.Event{}

	info.Labels = make(map[string]string)

	info.Service = os.Getenv("Http_Service")
	info.Owner = os.Getenv("Http_Owner")

	info.Repository = os.Getenv("Http_Repo")
	info.SHA = os.Getenv("Http_Sha")
	info.URL = os.Getenv("Http_Url")
	info.Image = os.Getenv("Http_Image")
	info.SCM = os.Getenv("Http_Scm")
	info.Private, _ = strconv.ParseBool(os.Getenv("Http_Private"))
	info.RepoURL = os.Getenv("Http_Repo_Url")

	if len(os.Getenv("Http_Owner_Id")) > 0 {
		info.OwnerID, _ = strconv.Atoi(os.Getenv("Http_Owner_Id"))
	}

	if len(os.Getenv("Http_Installation_id")) > 0 {
		info.InstallationID, err = strconv.Atoi(os.Getenv("Http_Installation_id"))
	}

	httpEnv := os.Getenv("Http_Env")
	envVars := make(map[string]string)

	if len(httpEnv) > 0 {
		unmarshalErr := json.Unmarshal([]byte(httpEnv), &envVars)

		if unmarshalErr == nil {
			info.Environment = envVars
		} else {
			log.Printf("Error un-marshaling env-vars map for function %s, %s", info.Service, unmarshalErr)
			info.Environment = make(map[string]string)
		}
	}

	httpLabels := os.Getenv("Http_Labels")
	labels := make(map[string]string)

	if len(httpLabels) > 0 {
		marshalErr := json.Unmarshal([]byte(httpLabels), &labels)
		if marshalErr == nil {
			info.Labels = labels
		} else {
			log.Printf("Error un-marshaling labels map for function %s, %s", info.Service, marshalErr)
			info.Labels = make(map[string]string)
		}
	}

	httpAnnotations := os.Getenv("Http_Annotations")
	annotations := make(map[string]string)

	if len(httpAnnotations) > 0 {
		marshalErr := json.Unmarshal([]byte(httpAnnotations), &annotations)
		if marshalErr == nil {
			info.Annotations = annotations
		} else {
			log.Printf("Error un-marshaling annotations map for function %s, %s", info.Service, marshalErr)
			info.Annotations = make(map[string]string)
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

func functionExists(deploy deployment, gatewayURL string) (bool, error) {

	r, _ := http.NewRequest(http.MethodGet, gatewayURL+"system/functions", nil)

	addAuthErr := sdk.AddBasicAuth(r)
	if addAuthErr != nil {
		log.Printf("Basic auth error %s", addAuthErr)
	}

	res, err := http.DefaultClient.Do(r)

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

func deployFunction(deploy deployment, gatewayURL string) (string, error) {
	exists, err := functionExists(deploy, gatewayURL)

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

	res, err = http.DefaultClient.Do(httpReq)

	if err != nil {
		log.Printf("error %s to system/functions %s", method, err)
		return "", err
	}

	defer res.Body.Close()

	log.Printf("Deploy status [%s] - %d", method, res.StatusCode)

	buildStatus, _ := ioutil.ReadAll(res.Body)

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return "", fmt.Errorf("http status code %d, error: %s", res.StatusCode, string(buildStatus))
	}

	return string(buildStatus), err
}

func enableStatusReporting() bool {
	return os.Getenv("report_status") == "true"
}

func reportStatus(status *sdk.Status, SCM string) error {
	if SCM == GitHub {
		reportGitHubStatus(status)
	} else if SCM == GitLab {
		reportGitLabStatus(status)
	} else {
		return fmt.Errorf("non-supported SCM: %s", SCM)
	}
	return nil
}

func reportGitHubStatus(status *sdk.Status) {

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
	Service                string
	Image                  string
	Network                string
	Labels                 map[string]string  `json:"labels"`
	Limits                 *FunctionResources `json:"limits,omitempty"`
	Requests               *FunctionResources `json:"requests,omitempty"`
	EnvVars                map[string]string  `json:"envVars"` // EnvVars provides overrides for functions.
	Secrets                []string           `json:"secrets"`
	ReadOnlyRootFilesystem bool               `json:"readOnlyRootFilesystem"`
	RegistryAuth           string             `json:"registryAuth"`
	Annotations            map[string]string  `json:"annotations"`
}

type Limits struct {
	Memory string
	CPU    string
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

// getCPULimit gives the CPU limit in millis if using Kubernetes
// for other orchestrators Available is set to false in the
// returned struct
func getCPULimit() CPULimits {
	var available bool

	kubernetesPort := "KUBERNETES_SERVICE_PORT"
	limit := ""
	requests := ""

	if _, exists := os.LookupEnv(kubernetesPort); exists {

		if val, ok := os.LookupEnv("function_cpu_limit_milli"); ok && len(val) > 0 {
			limit = fmt.Sprintf("%sm", val)
		}
		if val, ok := os.LookupEnv("function_cpu_requests_milli"); ok && len(val) > 0 {
			requests = fmt.Sprintf("%sm", val)
		}

		available = len(limit) > 0 || len(requests) > 0
	}

	return CPULimits{
		Available: available,
		Limit:     limit,
		Requests:  requests,
	}
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

	res, resErr := http.DefaultClient.Do(req)
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
	status.CommitStatuses = make(map[string]sdk.CommitStatus)
}

func buildBranch() string {
	branch := os.Getenv("build_branch")
	if branch == "" {
		return "master"
	}
	return branch
}

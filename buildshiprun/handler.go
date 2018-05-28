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
	"github.com/alexellis/derek/auth"
	"github.com/google/go-github/github"
)

const (
	defaultPrivateKeyName = "private_key.pem"
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

	r, _ := http.NewRequest(http.MethodPost, builderURL+"build", reader)
	r.Header.Set("Content-Type", "application/octet-stream")

	res, err := c.Do(r)

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

	serviceValue := ""

	log.Printf("buildshiprun: image '%s'\n", imageName)

	if strings.Contains(imageName, "exit status") == true {
		msg := "Unable to build image, check builder logs"
		status.AddStatus(sdk.Failure, msg, sdk.FunctionContext(event.Service))
		reportStatus(status)
		log.Fatal(msg)
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
		}

		log.Println(result)
	}

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
		fmt.Printf("failed to report status, error: %s", reportErr.Error())
	}

	if token == "" {
		fmt.Printf("failed to report status %v, error: authentication failed Invalid token\n", repoStatus)
		return
	}

	client := auth.MakeClient(ctx, token)

	_, _, apiErr := client.Repositories.CreateStatus(ctx, event.owner, event.repository, event.sha, repoStatus)
	if apiErr != nil {
		fmt.Printf("failed to report status %v, error: %s\n", repoStatus, apiErr.Error())
		return
	}
}

func getPrivateKey() string {
	// we are taking the secrets name from the env, by default it is fixed
	// to private_key.pem.
	// Although user can make the secret with a specific name and provide
	// it in the stack.yaml and also specify the secret name in github.yml
	privateKeyName := os.Getenv("private_key")
	if privateKeyName == "" {
		privateKeyName = defaultPrivateKeyName
	}
	privateKeyPath := "/run/secrets/" + privateKeyName
	return privateKeyPath
}

func buildStatus(status string, desc string, context string, url string) *github.RepoStatus {
	return &github.RepoStatus{State: &status, TargetURL: &url, Description: &desc, Context: &context}
}

func getImageName(repositoryURL, pushRepositoryURL, imageName string) string {

	return strings.Replace(imageName, pushRepositoryURL, repositoryURL, 1)

	// return repositoryURL + imageName[strings.Index(imageName, "/"):]
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

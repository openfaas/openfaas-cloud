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
)

// Handle a build / deploy request - returns empty string for an error
func Handle(req []byte) string {

	c := &http.Client{}

	builderURL := os.Getenv("builder_url")

	reader := bytes.NewBuffer(req)
	res, err := http.Post(builderURL+"build", "application/octet-stream", reader)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	defer res.Body.Close()

	buildStatus, _ := ioutil.ReadAll(res.Body)
	imageName := strings.TrimSpace(string(buildStatus))

	repositoryURL := os.Getenv("repository_url")

	if len(repositoryURL) == 0 {
		fmt.Fprintf(os.Stderr, "repository_url env-var not set")
		os.Exit(1)
	}

	if len(imageName) > 0 {
		gatewayURL := os.Getenv("gateway_url")

		service := os.Getenv("Http_Service")
		owner := os.Getenv("Http_Owner")
		repo := os.Getenv("Http_Repo")

		// Replace image name for "localhost" for deployment
		imageName = repositoryURL + imageName[strings.Index(imageName, ":"):]

		serviceValue := fmt.Sprintf("%s-%s", owner, service)

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
				"Git-Owner":      owner,
				"Git-Repo":       repo,
				"Git-DeployTime": time.Now().String(),
			},
			Limits: Limits{
				Memory: defaultMemoryLimit,
			},
		}

		result, err := deployFunction(deploy, gatewayURL, c)

		if err != nil {
			log.Fatal(err.Error())
		}

		log.Println(result)
	}

	return fmt.Sprintf("buildStatus %s %s %s", buildStatus, imageName, res.Status)
}

func functionExists(deploy deployment, gatewayURL string, c *http.Client) (bool, error) {

	res, err := http.Get(gatewayURL + "system/functions")

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

type deployment struct {
	Service string
	Image   string
	Network string
	Labels  map[string]string
	Limits  Limits
}

type Limits struct {
	Memory string
}

type function struct {
	Name string
}

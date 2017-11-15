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

// Handle a serverless request
func Handle(req []byte) string {
	reader := bytes.NewBuffer(req)
	res, err := http.Post("http://of-builder:8080/build", "application/octet-stream", reader)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	defer res.Body.Close()

	buildStatus, _ := ioutil.ReadAll(res.Body)
	imageName := strings.TrimSpace(string(buildStatus))

	if len(imageName) > 0 {
		service := os.Getenv("Http_Service")
		owner := os.Getenv("Http_Owner")
		repo := os.Getenv("Http_Repo")

		// Replace image name for local-host for deployment
		imageName = "127.0.0.1" + imageName[strings.Index(imageName, ":"):]

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
				"Git-Owner":      owner,
				"Git-Repo":       repo,
				"Git-DeployTime": time.Now().String(),
			},
			Limits: Limits{
				Memory: defaultMemoryLimit,
			},
		}

		result, err := deployFunction(deploy)
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Println(result)
	}

	return fmt.Sprintf("buildStatus %s %s %s", buildStatus, imageName, res.Status)
}

func functionExists(deploy deployment) (bool, error) {
	res, err := http.Get("http://gateway:8080/system/functions")

	if err != nil {
		fmt.Println(err)
		return false, err
	}

	defer res.Body.Close()

	fmt.Println("Deploy status: " + res.Status)
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

func deployFunction(deploy deployment) (string, error) {
	exists, err := functionExists(deploy)

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

	httpReq, err = http.NewRequest(method, "http://gateway:8080/system/functions", reader)
	c := http.Client{}
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

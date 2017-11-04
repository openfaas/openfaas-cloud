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
)

// Handle a serverless request
func Handle(req []byte) string {
	reader := bytes.NewBuffer(req)
	res, err := http.Post("http://of-builder:8080/build", "binary/octet-stream", reader)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	defer res.Body.Close()

	buildStatus, _ := ioutil.ReadAll(res.Body)

	imageName := strings.TrimSpace(string(buildStatus))
	imageName = "127.0.0.1" + imageName[strings.Index(imageName, ":"):]

	if len(imageName) > 0 {

		deploy := deployment{
			Service: os.Getenv("Http_Service"),
			Image:   imageName,
		}

		result, err := deployFunction(deploy)
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Println(result)
	}

	return fmt.Sprintf("buildStatus %s", buildStatus)
}

func deployFunction(deploy deployment) (string, error) {
	bytesOut, _ := json.Marshal(deploy)
	reader := bytes.NewBuffer(bytesOut)

	fmt.Println("Deploying: " + deploy.Image + " as " + deploy.Service)
	res, err := http.Post("http://gateway:8080/system/functions", "application/json", reader)
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
}

package function

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

	parts := strings.Split(strings.TrimSpace(string(buildStatus)), " ")
	if len(parts) == 2 {

		deploy := deployment{
			Service: os.Getenv("Http_Service"),
			Image:   parts[1],
		}

		deployFunction(deploy)
	}

	return fmt.Sprintf("buildStatus %s", buildStatus)
}

func deployFunction(deploy deployment) (string, error) {
	bytesOut, _ := json.Marshal(deploy)
	reader := bytes.NewBuffer(bytesOut)

	res, err := http.Post("http://gateway:8080/system/functions", "application/json", reader)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	defer res.Body.Close()

	buildStatus, _ := ioutil.ReadAll(res.Body)
	return string(buildStatus), err
}

type deployment struct {
	Service string
	Image   string
}

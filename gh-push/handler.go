package function

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// Handle a serverless request
func Handle(req []byte) string {

	event := os.Getenv("Http_X_Github_Event")

	if event == "push" {
		pushEvent := PushEvent{}
		err := json.Unmarshal(req, &pushEvent)
		if err != nil {
			return err.Error()
		}

		body, _ := json.Marshal(pushEvent)

		c := http.Client{}
		bodyReader := bytes.NewBuffer(body)
		httpReq, _ := http.NewRequest(http.MethodPost, "http://gateway:8080/async-function/git-tar", bodyReader)
		res, reqErr := c.Do(httpReq)
		if reqErr != nil {
			return reqErr.Error()
		}

		fmt.Println("Tar - ", res.StatusCode)

		return fmt.Sprintf("Got a push - %s\n", pushEvent)
	}

	return "I can't handle event: " + event
}

type PushEvent struct {
	Repository struct {
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		CloneURL string `json:"clone_url"`
	}
	AfterCommitID string `json:"after"`
}

func Init() {

}

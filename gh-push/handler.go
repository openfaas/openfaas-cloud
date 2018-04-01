package function

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/alexellis/hmac"
)

// Handle a serverless request
func Handle(req []byte) string {

	event := os.Getenv("Http_X_Github_Event")

	if event == "push" {
		xHubSignature := os.Getenv("Http_X_Hub_Signature")

		shouldValidate := os.Getenv("validate_hmac")
		if len(shouldValidate) > 0 && (shouldValidate == "1" || shouldValidate == "true") {
			validateErr := hmac.Validate(req, xHubSignature, os.Getenv("github_webhook_secret"))
			if validateErr != nil {
				log.Fatal(validateErr)
			}
		}

		pushEvent := PushEvent{}
		err := json.Unmarshal(req, &pushEvent)
		if err != nil {
			return err.Error()
		}
		statusCode, postErr := postEvent(pushEvent)
		if postErr != nil {
			return postErr.Error()
		}

		return fmt.Sprintf("Push - %s, git-tar status: %d\n", pushEvent, statusCode)
	}

	return fmt.Sprintf("gh-push cannot handle event: %s", event)
}

func postEvent(pushEvent PushEvent) (int, error) {
	gatewayURL := os.Getenv("gateway_url")

	body, _ := json.Marshal(pushEvent)

	c := http.Client{}
	bodyReader := bytes.NewBuffer(body)
	httpReq, _ := http.NewRequest(http.MethodPost, gatewayURL+"async-function/git-tar", bodyReader)
	res, reqErr := c.Do(httpReq)

	if reqErr != nil {
		return http.StatusServiceUnavailable, reqErr
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	return res.StatusCode, nil
}

// PushEvent as received from GitHub
type PushEvent struct {
	Repository struct {
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		CloneURL string `json:"clone_url"`
		Owner    struct {
			Login string `json:"login"`
			Email string `json:"email"`
		} `json:"owner"`
	}
	AfterCommitID string `json:"after"`
}

func Init() {

}

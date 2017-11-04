package function

import (
	"encoding/json"
	"fmt"
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

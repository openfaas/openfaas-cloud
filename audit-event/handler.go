package function

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/openfaas/openfaas-cloud/sdk"
)

type SlackMsg struct {
	Text string `json:"text"`
}

// Handle a serverless request
func Handle(req []byte) string {

	event := sdk.AuditEvent{}

	json.Unmarshal(req, &event)

	log.Printf("Event: %s", req)

	slackURL := os.Getenv("slack_url")
	if len(slackURL) > 0 {
		reader, encapsulateErr := encapsulateSlackReq(event)
		if encapsulateErr != nil {
			log.Panic(encapsulateErr)
		}

		res, err := http.Post(slackURL, "application/json", reader)
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Posted to Slack: ", res.Status)
		}
	}

	return fmt.Sprintf("audit-event: done")
}

func encapsulateSlackReq(event sdk.AuditEvent) (io.Reader, error) {
	msg := SlackMsg{Text: fmt.Sprintf("[%s] %s/%s: '%s'", event.Source, event.Owner, event.Repo, event.Message)}

	bytesOut, marshalErr := json.Marshal(msg)
	if marshalErr != nil {
		return nil, marshalErr
	}

	return bytes.NewReader(bytesOut), nil
}

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

// SlackMessage encapsulates a message for the Slack
// incoming webhook API
type SlackMessage struct {
	Text string `json:"text"`
}

// Handle collects events from other functions for auditing. These can
// be connected to a Slack webhook URL or the function can be swapped
// for the echo  function for storage in container logs.
func Handle(req []byte) string {

	event := sdk.AuditEvent{}

	json.Unmarshal(req, &event)

	log.Printf("Event: %s", req)

	if slackURL, ok := os.LookupEnv("slack_url"); ok && len(slackURL) > 0 {
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
	msg := SlackMessage{
		Text: fmt.Sprintf("[%s] %s/%s: '%s'",
			event.Source,
			event.Owner,
			event.Repo,
			event.Message)}

	bytesOut, marshalErr := json.Marshal(msg)
	if marshalErr != nil {
		return nil, marshalErr
	}

	return bytes.NewReader(bytesOut), nil
}

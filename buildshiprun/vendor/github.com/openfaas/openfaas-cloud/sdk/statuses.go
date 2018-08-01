package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

// status constant
const (
	Success = "success"
	Failure = "failure"
	Pending = "pending"
)

// context constant
const (
	Deploy = "%s"
	Stack  = "stack-deploy"
)

const authTokenPattern = "^[A-Za-z0-9-_.]*"

var validToken = regexp.MustCompile(authTokenPattern)

type CommitStatus struct {
	Status      string `json:"status"`
	Description string `json:"description"`
	Context     string `json:"context"`
}

// Status to post github-status to git-status function
type Status struct {
	CommitStatuses map[string]CommitStatus `json:"commit-statuses"`
	EventInfo      Event                   `json:"event"`
	AuthToken      string                  `json:"auth-token"`
}

// builds a status
func BuildStatus(event *Event, token string) *Status {

	status := Status{}
	status.EventInfo = *event
	status.CommitStatuses = make(map[string]CommitStatus)
	status.AuthToken = token

	return &status
}

// adds a commit status into a status
// a status can contain multiple commit status
func (status *Status) AddStatus(state string, desc string, context string) {
	if status.CommitStatuses == nil {
		status.CommitStatuses = make(map[string]CommitStatus)
	}
	// the status.CommitStatuses isn map so that it replace the old commit status for same context
	status.CommitStatuses[context] = CommitStatus{Status: state, Description: desc, Context: context}
}

// unmardhal a status
func UnmarshalStatus(data []byte) (*Status, error) {
	status := Status{}
	err := json.Unmarshal(data, &status)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// masrshal a status
func (status *Status) Marshal() ([]byte, error) {
	return json.Marshal(status)
}

// send a status update to git-status function
func (status *Status) Report(gateway string) (string, error) {
	body, _ := status.Marshal()

	c := http.Client{}
	bodyReader := bytes.NewBuffer(body)
	httpReq, _ := http.NewRequest(http.MethodPost, gateway+"function/github-status", bodyReader)

	res, err := c.Do(httpReq)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()
	resData, _ := ioutil.ReadAll(res.Body)
	if resData == nil {
		return "", fmt.Errorf("empty token received")
	}

	token := string(resData)
	token = strings.TrimSuffix(token, "\n")
	if !ValidToken(token) {
		log.Printf(`invalid auth token received, token : ( %s ),
make sure combine_output is disabled for git-status`, token)
		status.AuthToken = ""
	} else {
		status.AuthToken = token
	}
	// reset old status
	status.CommitStatuses = make(map[string]CommitStatus)

	return string(token), nil
}

func ValidToken(token string) bool {
	match := validToken.FindString(token)
	// token should be the whole string
	if len(match) == len(token) {
		return true
	}
	return false
}

// build a github context for a function
// a context for function build can be created as:
//   sdk.GetContext(functionName)
func FunctionContext(function string) string {
	return fmt.Sprintf(Deploy, function)
}

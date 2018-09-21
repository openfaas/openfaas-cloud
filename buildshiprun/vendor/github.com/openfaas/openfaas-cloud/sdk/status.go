package sdk

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"

	hmac "github.com/alexellis/hmac"
)

// github status constant
const (
	StatusSuccess = "success"
	StatusFailure = "failure"
	StatusPending = "pending"
)

// context constant
const (
	FunctionContext = "%s"
	StackContext    = "stack-deploy"
	EmptyAuthToken  = ""
	tokenKey        = "token"
)

const authTokenPattern = "^[A-Za-z0-9-_.]*"

var validToken = regexp.MustCompile(authTokenPattern)

// CommitStatus to be written to GitHub/GitLab
type CommitStatus struct {
	Status      string `json:"status"`
	Description string `json:"description"`
	Context     string `json:"context"`
}

// Status to post status to github-status function
type Status struct {
	CommitStatuses map[string]CommitStatus `json:"commit-statuses"`
	EventInfo      Event                   `json:"event"`
	AuthToken      string                  `json:"auth-token"`
}

// BuildStatus constructs a status object from event
func BuildStatus(event *Event, token string) *Status {
	return &Status{
		EventInfo:      *event,
		CommitStatuses: make(map[string]CommitStatus),
		AuthToken:      token,
	}
}

// UnmarshalStatus unmarshals a status object from json
func UnmarshalStatus(data []byte) (*Status, error) {
	status := Status{}
	err := json.Unmarshal(data, &status)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// Clear removes any statuses which have been added
func (status *Status) Clear() {
	status.CommitStatuses = make(map[string]CommitStatus)
}

// AddStatus adds a commit status into a status object
// a status can contain multiple commit status
func (status *Status) AddStatus(state string, desc string, context string) {

	// TODO: AE - don't think these lines are required
	if status.CommitStatuses == nil {
		status.CommitStatuses = make(map[string]CommitStatus)
	}

	// the status.CommitStatuses is a map hashed against the context
	// it replace the old commit status if added for same context
	status.CommitStatuses[context] = CommitStatus{Status: state, Description: desc, Context: context}
}

// Marshal marshals a status into json
func (status *Status) Marshal() ([]byte, error) {
	return json.Marshal(status)
}

// ValidToken check if a token is in valid format
func ValidToken(token string) bool {
	match := validToken.FindString(token)
	// token should be the whole string
	if len(match) == len(token) {
		return true
	}
	return false
}

// MarshalToken marshal a token into json i.e. {"token": "auth_token_value"}
func MarshalToken(token string) string {
	marshalToken, _ := json.Marshal(map[string]string{tokenKey: token})
	return string(marshalToken)
}

// UnmarshalToken unmarshal a token and validate
func UnmarshalToken(data []byte) (string, error) {
	tokenMap := make(map[string]string)

	err := json.Unmarshal(data, &tokenMap)
	if err != nil {
		return EmptyAuthToken, fmt.Errorf(`invalid auth token format received: %s. error: %s, make sure combine_output is disabled for github-status`, data, err)
	}

	token := tokenMap[tokenKey]
	if !ValidToken(token) {
		return EmptyAuthToken, fmt.Errorf(`invalid auth token received, token : ( %s ),
make sure combine_output is disabled for github-status`, token)
	}
	return token, nil
}

// Report send a status update to github-status function
func (status *Status) Report(gateway string, payloadSecret string) (string, error) {
	body, _ := status.Marshal()

	c := http.Client{}
	bodyReader := bytes.NewBuffer(body)
	httpReq, _ := http.NewRequest(http.MethodPost, gateway+"function/github-status", bodyReader)

	if len(payloadSecret) > 0 {
		digest := hmac.Sign(body, []byte(payloadSecret))
		httpReq.Header.Add(CloudSignatureHeader, "sha1="+hex.EncodeToString(digest))
	}

	res, err := c.Do(httpReq)
	if err != nil {
		return "", err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	resData, readErr := ioutil.ReadAll(res.Body)
	if resData == nil || readErr != nil {
		return "", fmt.Errorf("failed to read response from github-status")
	}

	if res.StatusCode != http.StatusAccepted && res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to call github-status, invalid status: %s", res.Status)
	}

	status.AuthToken, err = UnmarshalToken(resData)
	if err != nil {
		log.Printf(err.Error())
	}

	// reset old status
	status.CommitStatuses = make(map[string]CommitStatus)

	return status.AuthToken, nil
}

// BuildFunctionContext build a github context for a function
//                      Example:
//                        sdk.BuildFunctionContext(functionName)
func BuildFunctionContext(function string) string {
	return fmt.Sprintf(FunctionContext, function)
}

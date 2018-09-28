package function

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/openfaas/openfaas-cloud/sdk"
)

func TestBuildURLWithoutPrettyURL_WithSlash(t *testing.T) {
	os.Setenv("gateway_public_url", "http://localhost:8080")
	os.Setenv("gateway_pretty_url", "")

	event := &sdk.Event{
		Owner:   "alexellis",
		Service: "tester",
	}

	val := buildPublicStatusURL("success", sdk.BuildFunctionContext(event.Service), event)
	want := "http://localhost:8080/function/alexellis-tester"

	if val != want {
		t.Errorf("building PublicURL: want %s, got %s", want, val)
		t.Fail()
	}
}

func TestBuildURLWithoutPrettyURL_WithOutSlash(t *testing.T) {
	os.Setenv("gateway_public_url", "http://localhost:8080")
	os.Setenv("gateway_pretty_url", "")

	event := &sdk.Event{
		Owner:   "alexellis",
		Service: "tester",
	}

	val := buildPublicStatusURL("success", sdk.BuildFunctionContext(event.Service), event)
	want := "http://localhost:8080/function/alexellis-tester"

	if val != want {
		t.Errorf("building PublicURL: want %s, got %s", want, val)
		t.Fail()
	}
}

func TestBuildURLWithPrettyURL(t *testing.T) {
	os.Setenv("gateway_public_url", "http://localhost:8080")
	os.Setenv("gateway_pretty_url", "https://user.openfaas-cloud.com/function")

	event := &sdk.Event{
		Owner:   "alexellis",
		Service: "tester",
	}

	val := buildPublicStatusURL("success", sdk.BuildFunctionContext(event.Service), event)
	want := "https://alexellis.openfaas-cloud.com/tester"

	if val != want {
		t.Errorf("building PublicURL: want %s, got %s", want, val)
		t.Fail()
	}
}

func TestBuildURLWithUndefinedStatusGivesOriginalURL(t *testing.T) {

	event := &sdk.Event{
		Owner:   "alexellis",
		Service: "tester",
		URL:     "http://original-value.local",
	}

	val := buildPublicStatusURL("not-supported", sdk.BuildFunctionContext(event.Service), event)
	want := event.URL

	if val != want {
		t.Errorf("building PublicURL: want %s, got %s", want, val)
		t.Fail()
	}
}

func TestToken(t *testing.T) {
	testTokens := []struct {
		title         string
		token         string
		validToken    bool
		expectedError bool
		expectedValue string
	}{
		{
			"Token with space",
			"token with space",
			false,
			true,
			sdk.EmptyAuthToken,
		},
		{
			"Token with special char",
			"token$With@special=Char",
			false,
			true,
			sdk.EmptyAuthToken,
		},
		{
			"Valid Token",
			"v1.afbce39asdasd8be30123317cef123321ae991cf40f7",
			true,
			false,
			"v1.afbce39asdasd8be30123317cef123321ae991cf40f7",
		},
		{
			"Token invalid string",
			"token=v1.afbce39asdasd8be30123",
			false,
			true,
			sdk.EmptyAuthToken,
		},
		{
			"Empty Token",
			" ",
			false,
			true,
			sdk.EmptyAuthToken,
		},
	}
	for _, test := range testTokens {
		t.Run(test.title, func(t *testing.T) {
			// Test token validity
			if sdk.ValidToken(test.token) != test.validToken {
				t.Errorf("validating token %s: want %v, got %v", test.token, test.validToken, !test.validToken)
			}
			// Check Token and Unmarshal success and failure cases
			marshaledToken := sdk.MarshalToken(test.token)
			unmarshaledToken, err := sdk.UnmarshalToken([]byte(marshaledToken))
			if unmarshaledToken != test.expectedValue {
				t.Errorf("validating token %s: want %s, got %s", test.token, test.expectedValue, unmarshaledToken)
			}
			// Check if expected error happened
			if (err != nil) != test.expectedError {
				expect := "expected error"
				if !test.expectedError {
					expect = "expected no error"
				}
				t.Errorf("validating invalid token error `%s`: %v, got: %v", test.token, expect, err)
			}
		})
	}
}

func TestTokenUnmarshalError(t *testing.T) {
	testTokens := []string{
		"{\"token\":\"token with space\"",
		"token=\"@$5^Char\"",
		" ",
		"token=",
		"{\"token\"=\"\"}",
	}
	for _, invalidToken := range testTokens {
		unmarshaledToken, err := sdk.UnmarshalToken([]byte(invalidToken))
		if err == nil {
			t.Errorf("validating invalid token encoding `%s`: expected error, got token %s", token, unmarshaledToken)
		}
	}
}

func TestStatusCreation(t *testing.T) {
	event := &sdk.Event{
		Owner:   "alexellis",
		Service: "tester",
		URL:     "http://original-value.local",
	}
	status := sdk.BuildStatus(event, sdk.EmptyAuthToken)
	if status == nil {
		t.Errorf("validating status creation: got %v", status)
	}
	if status.CommitStatuses == nil {
		t.Errorf("validating commitstatuses: got %v", status.CommitStatuses)
	}
}

func TestStatusCreationWithToken(t *testing.T) {
	event := &sdk.Event{
		Owner:   "alexellis",
		Service: "tester",
		URL:     "http://original-value.local",
	}
	status := sdk.BuildStatus(event, "dummyauthtoken")
	if status.AuthToken != "dummyauthtoken" {
		t.Errorf("validating status token: want %v got %v", "dummyauthtoken", status.AuthToken)
	}
}

func TestBuildFunctionContext(t *testing.T) {
	function := "my_func"
	if sdk.BuildFunctionContext(function) != function {
		t.Errorf("validating function context: want %v got %v", function, sdk.BuildFunctionContext(function))
	}
}

func TestStatusAddition(t *testing.T) {
	event := &sdk.Event{
		Owner:   "alexellis",
		Service: "tester",
		URL:     "http://original-value.local",
	}
	status := sdk.BuildStatus(event, sdk.EmptyAuthToken)
	status.AddStatus(sdk.StatusPending, "description stack", sdk.StackContext)
	status.AddStatus(sdk.StatusSuccess, "description func", sdk.BuildFunctionContext("func"))

	commitStatus, ok := status.CommitStatuses[sdk.StackContext]
	if !ok {
		t.Errorf("validating commit status addition: got %v", commitStatus)
	}
	if commitStatus.Status != sdk.StatusPending {
		t.Errorf("validating commit status state: want %v got %v", sdk.StatusPending, commitStatus.Status)
	}
	if commitStatus.Description != "description stack" {
		t.Errorf("validating commit status description: want %v got %v", "description stack", commitStatus.Description)
	}
	if commitStatus.Context != sdk.StackContext {
		t.Errorf("validating commit status context: want %v got %v", sdk.StackContext, commitStatus.Context)
	}

	commitStatus, ok = status.CommitStatuses[sdk.BuildFunctionContext("func")]
	if !ok {
		t.Errorf("validating commit status addition: got %v", commitStatus)
	}
	if commitStatus.Status != sdk.StatusSuccess {
		t.Errorf("validating commit status state: want %v got %v", sdk.StatusSuccess, commitStatus.Status)
	}
	if commitStatus.Description != "description func" {
		t.Errorf("validating commit status description: want %v got %v", "description func", commitStatus.Description)
	}
	if commitStatus.Context != sdk.BuildFunctionContext("func") {
		t.Errorf("validating commit status context: want %v got %v", sdk.BuildFunctionContext("func"), commitStatus.Context)
	}
}

func TestStatusOverwriteForSameContext(t *testing.T) {
	event := &sdk.Event{
		Owner:   "alexellis",
		Service: "tester",
		URL:     "http://original-value.local",
	}
	status := sdk.BuildStatus(event, sdk.EmptyAuthToken)
	status.AddStatus(sdk.StatusPending, "description stack pending", sdk.StackContext)
	status.AddStatus(sdk.StatusPending, "description func pending", sdk.BuildFunctionContext("func"))
	status.AddStatus(sdk.StatusSuccess, "description stack success", sdk.StackContext)
	status.AddStatus(sdk.StatusFailure, "description func failure", sdk.BuildFunctionContext("func"))

	commitStatus, ok := status.CommitStatuses[sdk.StackContext]
	if !ok {
		t.Errorf("validating overwritten commit status addition: got %v", commitStatus)
	}
	if commitStatus.Status != sdk.StatusSuccess {
		t.Errorf("validating overwritten commit status state: want %v got %v", sdk.StatusSuccess, commitStatus.Status)
	}
	if commitStatus.Description != "description stack success" {
		t.Errorf("validating overwritten commit status description: want %v got %v", "description stack success", commitStatus.Description)
	}
	if commitStatus.Context != sdk.StackContext {
		t.Errorf("validating overwritten commit status context: want %v got %v", sdk.StackContext, commitStatus.Context)
	}

	commitStatus, ok = status.CommitStatuses[sdk.BuildFunctionContext("func")]
	if !ok {
		t.Errorf("validating overwritten commit status addition: got %v", commitStatus)
	}
	if commitStatus.Status != sdk.StatusFailure {
		t.Errorf("validating overwritten commit status state: want %v got %v", sdk.StatusFailure, commitStatus.Status)
	}
	if commitStatus.Description != "description func failure" {
		t.Errorf("validating overwritten commit status description: want %v got %v", "description func failure", commitStatus.Description)
	}
	if commitStatus.Context != sdk.BuildFunctionContext("func") {
		t.Errorf("validating overwritten commit status context: want %v got %v", sdk.BuildFunctionContext("func"), commitStatus.Context)
	}
}

func TestStatusEncodingDecoding(t *testing.T) {
	event := &sdk.Event{
		Owner:   "alexellis",
		Service: "tester",
		URL:     "http://original-value.local",
	}
	status := sdk.BuildStatus(event, "dummyauthtoken")
	status.AddStatus(sdk.StatusPending, "description stack", sdk.StackContext)
	status.AddStatus(sdk.StatusSuccess, "description func", sdk.BuildFunctionContext("func"))

	data, err := status.Marshal()
	if err != nil {
		t.Errorf("validating status encoding failed: got %v", err)
	}

	decodedStatus, err := sdk.UnmarshalStatus(data)
	if err != nil {
		t.Errorf("validating status decoding failed: got %v", err)
	}

	commitStatus, ok := decodedStatus.CommitStatuses[sdk.StackContext]
	if !ok {
		t.Errorf("validating decoded commit status addition: got %v", commitStatus)
	}
	if commitStatus.Status != sdk.StatusPending {
		t.Errorf("validating decoded commit status state: want %v got %v", sdk.StatusPending, commitStatus.Status)
	}
	if commitStatus.Description != "description stack" {
		t.Errorf("validating decoded commit status description: want %v got %v", "description stack", commitStatus.Description)
	}
	if commitStatus.Context != sdk.StackContext {
		t.Errorf("validating decoded commit status context: want %v got %v", sdk.StackContext, commitStatus.Context)
	}

	commitStatus, ok = decodedStatus.CommitStatuses[sdk.BuildFunctionContext("func")]
	if !ok {
		t.Errorf("validating decoded commit status addition: got %v", commitStatus)
	}
	if commitStatus.Status != sdk.StatusSuccess {
		t.Errorf("validating decoded commit status state: want %v got %v", sdk.StatusSuccess, commitStatus.Status)
	}
	if commitStatus.Description != "description func" {
		t.Errorf("validating decoded commit status description: want %v got %v", "description func", commitStatus.Description)
	}
	if commitStatus.Context != sdk.BuildFunctionContext("func") {
		t.Errorf("validating decoded commit status context: want %v got %v", sdk.BuildFunctionContext("func"), commitStatus.Context)
	}
}

func TestStatusReportFailure(t *testing.T) {
	event := &sdk.Event{
		Owner:   "alexellis",
		Service: "tester",
		URL:     "http://original-value.local",
	}
	status := sdk.BuildStatus(event, sdk.EmptyAuthToken)
	status.AddStatus(sdk.StatusPending, "description stack", sdk.StackContext)

	gateway := "invalid:8080/"
	token, err := status.Report(gateway, "")
	if token != "" {
		t.Errorf("validating report failure token: want %v got %v", "", token)
	}

	if err == nil {
		t.Errorf("validating report failure: got %v", nil)
	}
}

func TestPrivateKey(t *testing.T) {
	os.Setenv("private_key_filename", "github_key")
	os.Setenv("secret_mount_path", "/function/secrets")

	expectedPath := filepath.Join("/function/secrets/", "github_key")
	privateKey := sdk.GetPrivateKeyPath()
	if privateKey != expectedPath {
		t.Errorf("validating private key path: watch %v got %v", expectedPath, privateKey)
	}
}

func TestGetCheckRunTitle(t *testing.T) {
	status := &sdk.CommitStatus{
		Context:     sdk.StackContext,
		Description: "deployed: kvuchkov-hello-go",
		Status:      sdk.StatusSuccess,
	}
	title := getCheckRunTitle(status)
	if *title != "Deploy to OpenFaaS" {
		t.Fatalf("Expected %s but got %s", "Deploy to OpenFaaS", *title)
	}

	status.Context = sdk.BuildFunctionContext("hello-go")
	title = getCheckRunTitle(status)
	if *title != "Build hello-go" {
		t.Fatalf("Expected %s but got %s", "Build hello-go", *title)
	}
}

func TestGetCheckRunStatus(t *testing.T) {
	status := sdk.StatusFailure
	checkStatus := getCheckRunStatus(&status)
	if checkStatus != "completed" {
		t.Fatalf("Expected %s, got %s", "completed", checkStatus)
	}

	status = sdk.StatusSuccess
	checkStatus = getCheckRunStatus(&status)
	if checkStatus != "completed" {
		t.Fatalf("Expected %s, got %s", "completed", checkStatus)
	}

	status = sdk.StatusPending
	checkStatus = getCheckRunStatus(&status)
	if checkStatus != "queued" {
		t.Fatalf("Expected %s, got %s", "queued", checkStatus)
	}
}

// Test_formatLog tests formatting for the GitHub Checks API
func Test_formatLog(t *testing.T) {
	tests := []struct {
		title     string
		maxLength int
		rawLog    string
		wantLog   string
	}{
		{
			title:     "Log length is valid",
			maxLength: 500,
			rawLog:    "apk add --no-cache curl the container builder exported the image in 10s",
			wantLog:   "\n```shell\napk add --no-cache curl the container builder exported the image in 10s\n```\n",
		},
		{
			title:     "Log length is too short for warning, but needs truncating",
			maxLength: 20,
			rawLog:    "the container builder exported the image in 10s",
			wantLog:   "\n```shell\nted the image in 10s\n```\n",
		},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			got := formatLog(test.rawLog, test.maxLength)
			if got != test.wantLog {
				t.Errorf("want:\n`%q`\ngot:\n`%q`\n", test.wantLog, got)
			}
		})
	}

}

func Test_truncate(t *testing.T) {
	tests := []struct {
		title           string
		length          int
		message         string
		expectedMessage string
	}{
		{
			title:           "Exceeding 5 characters",
			length:          7,
			message:         "Some random string",
			expectedMessage: " string",
		},
		{
			title:           "Not exceeding 5 characters",
			length:          5,
			message:         "Some",
			expectedMessage: "Some",
		},
		{
			title:           "Right at 5 characters",
			length:          5,
			message:         "Some ",
			expectedMessage: "Some ",
		},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			message := truncate(test.length, test.message)
			if message != test.expectedMessage {
				t.Errorf("Expected message to be: `%s`, got: `%s`", message, test.expectedMessage)
			}
		})
	}
}

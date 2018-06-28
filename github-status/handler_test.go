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

	val := buildPublicStatusURL("success", sdk.FunctionContext(event.Service), event)
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

	val := buildPublicStatusURL("success", sdk.FunctionContext(event.Service), event)
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

	val := buildPublicStatusURL("success", sdk.FunctionContext(event.Service), event)
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

	val := buildPublicStatusURL("not-supported", sdk.FunctionContext(event.Service), event)
	want := event.URL

	if val != want {
		t.Errorf("building PublicURL: want %s, got %s", want, val)
		t.Fail()
	}
}

func TestTokenValidator(t *testing.T) {
	testTokens := map[string]bool{"token with space": false, "token$With@special=Char": false, "v1.afbce39asdasd8be30123317cef123321ae991cf40f7": true, "token=v1.afbce39asdasd8be30123": false, " ": false}
	for token, result := range testTokens {
		if sdk.ValidToken(token) != result {
			t.Errorf("validating token %s: want %v, got %v", token, result, !result)
		}
	}
}

func TestStatusCreation(t *testing.T) {
	event := &sdk.Event{
		Owner:   "alexellis",
		Service: "tester",
		URL:     "http://original-value.local",
	}
	status := sdk.BuildStatus(event, "")
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

func TestFunctionContext(t *testing.T) {
	function := "my_func"
	if sdk.FunctionContext(function) != function {
		t.Errorf("validating function context: want %v got %v", function, sdk.FunctionContext(function))
	}
}

func TestStatusAddition(t *testing.T) {
	event := &sdk.Event{
		Owner:   "alexellis",
		Service: "tester",
		URL:     "http://original-value.local",
	}
	status := sdk.BuildStatus(event, "")
	status.AddStatus(sdk.Pending, "description stack", sdk.Stack)
	status.AddStatus(sdk.Success, "description func", sdk.FunctionContext("func"))

	commitStatus, ok := status.CommitStatuses[sdk.Stack]
	if !ok {
		t.Errorf("validating commit status addition: got %v", commitStatus)
	}
	if commitStatus.Status != sdk.Pending {
		t.Errorf("validating commit status state: want %v got %v", sdk.Pending, commitStatus.Status)
	}
	if commitStatus.Description != "description stack" {
		t.Errorf("validating commit status description: want %v got %v", "description stack", commitStatus.Description)
	}
	if commitStatus.Context != sdk.Stack {
		t.Errorf("validating commit status context: want %v got %v", sdk.Stack, commitStatus.Context)
	}

	commitStatus, ok = status.CommitStatuses[sdk.FunctionContext("func")]
	if !ok {
		t.Errorf("validating commit status addition: got %v", commitStatus)
	}
	if commitStatus.Status != sdk.Success {
		t.Errorf("validating commit status state: want %v got %v", sdk.Success, commitStatus.Status)
	}
	if commitStatus.Description != "description func" {
		t.Errorf("validating commit status description: want %v got %v", "description func", commitStatus.Description)
	}
	if commitStatus.Context != sdk.FunctionContext("func") {
		t.Errorf("validating commit status context: want %v got %v", sdk.FunctionContext("func"), commitStatus.Context)
	}
}

func TestStatusOverwriteForSameContext(t *testing.T) {
	event := &sdk.Event{
		Owner:   "alexellis",
		Service: "tester",
		URL:     "http://original-value.local",
	}
	status := sdk.BuildStatus(event, "")
	status.AddStatus(sdk.Pending, "description stack pending", sdk.Stack)
	status.AddStatus(sdk.Pending, "description func pending", sdk.FunctionContext("func"))
	status.AddStatus(sdk.Success, "description stack success", sdk.Stack)
	status.AddStatus(sdk.Failure, "description func failure", sdk.FunctionContext("func"))

	commitStatus, ok := status.CommitStatuses[sdk.Stack]
	if !ok {
		t.Errorf("validating overwritten commit status addition: got %v", commitStatus)
	}
	if commitStatus.Status != sdk.Success {
		t.Errorf("validating overwritten commit status state: want %v got %v", sdk.Success, commitStatus.Status)
	}
	if commitStatus.Description != "description stack success" {
		t.Errorf("validating overwritten commit status description: want %v got %v", "description stack success", commitStatus.Description)
	}
	if commitStatus.Context != sdk.Stack {
		t.Errorf("validating overwritten commit status context: want %v got %v", sdk.Stack, commitStatus.Context)
	}

	commitStatus, ok = status.CommitStatuses[sdk.FunctionContext("func")]
	if !ok {
		t.Errorf("validating overwritten commit status addition: got %v", commitStatus)
	}
	if commitStatus.Status != sdk.Failure {
		t.Errorf("validating overwritten commit status state: want %v got %v", sdk.Failure, commitStatus.Status)
	}
	if commitStatus.Description != "description func failure" {
		t.Errorf("validating overwritten commit status description: want %v got %v", "description func failure", commitStatus.Description)
	}
	if commitStatus.Context != sdk.FunctionContext("func") {
		t.Errorf("validating overwritten commit status context: want %v got %v", sdk.FunctionContext("func"), commitStatus.Context)
	}
}

func TestStatusEncodingDecoding(t *testing.T) {
	event := &sdk.Event{
		Owner:   "alexellis",
		Service: "tester",
		URL:     "http://original-value.local",
	}
	status := sdk.BuildStatus(event, "dummyauthtoken")
	status.AddStatus(sdk.Pending, "description stack", sdk.Stack)
	status.AddStatus(sdk.Success, "description func", sdk.FunctionContext("func"))

	data, err := status.Marshal()
	if err != nil {
		t.Errorf("validating status encoding failed: got %v", err)
	}

	decodedStatus, err := sdk.UnmarshalStatus(data)
	if err != nil {
		t.Errorf("validating status decoding failed: got %v", err)
	}

	commitStatus, ok := decodedStatus.CommitStatuses[sdk.Stack]
	if !ok {
		t.Errorf("validating decoded commit status addition: got %v", commitStatus)
	}
	if commitStatus.Status != sdk.Pending {
		t.Errorf("validating decoded commit status state: want %v got %v", sdk.Pending, commitStatus.Status)
	}
	if commitStatus.Description != "description stack" {
		t.Errorf("validating decoded commit status description: want %v got %v", "description stack", commitStatus.Description)
	}
	if commitStatus.Context != sdk.Stack {
		t.Errorf("validating decoded commit status context: want %v got %v", sdk.Stack, commitStatus.Context)
	}

	commitStatus, ok = decodedStatus.CommitStatuses[sdk.FunctionContext("func")]
	if !ok {
		t.Errorf("validating decoded commit status addition: got %v", commitStatus)
	}
	if commitStatus.Status != sdk.Success {
		t.Errorf("validating decoded commit status state: want %v got %v", sdk.Success, commitStatus.Status)
	}
	if commitStatus.Description != "description func" {
		t.Errorf("validating decoded commit status description: want %v got %v", "description func", commitStatus.Description)
	}
	if commitStatus.Context != sdk.FunctionContext("func") {
		t.Errorf("validating decoded commit status context: want %v got %v", sdk.FunctionContext("func"), commitStatus.Context)
	}
}

func TestStatusReportFailure(t *testing.T) {
	event := &sdk.Event{
		Owner:   "alexellis",
		Service: "tester",
		URL:     "http://original-value.local",
	}
	status := sdk.BuildStatus(event, "")
	status.AddStatus(sdk.Pending, "description stack", sdk.Stack)

	gateway := "invalid:8080/"
	token, err := status.Report(gateway)
	if token != "" {
		t.Errorf("validating report failure token: want %v got %v", "", token)
	}

	if err == nil {
		t.Errorf("validating report failure: got %v", nil)
	}
}

func TestPrivateKey(t *testing.T) {
	os.Setenv("private_key", "github_key")
	os.Setenv("secret_mount_path", "/function/secrets")

	expectedPath := filepath.Join("/function/secrets/", "github_key")
	privateKey := getPrivateKey()
	if privateKey != expectedPath {
		t.Errorf("validating private key path: watch %v got %v", expectedPath, privateKey)
	}
}

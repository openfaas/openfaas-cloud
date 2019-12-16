package function

import (
	"github.com/openfaas/openfaas-cloud/sdk"
	"os"
	"testing"
)

func TestBuildURLWithoutPrettyURL_WithSlash(t *testing.T) {
	os.Setenv("gateway_public_url", "http://localhost:8080")
	os.Setenv("gateway_pretty_url", "")

	event := &sdk.Event{
		Owner:   "alexellis",
		Service: "tester",
	}

	got := buildPublicStatusURL("success", sdk.BuildFunctionContext(event.Service), event)
	want := "http://localhost:8080/function/alexellis-tester"

	if got != want {
		t.Errorf("building PublicURL: want %s, got %s", want, got)
		t.Fail()
	}
}

func TestBuildURLWithoutPrettyURL_WithSlashStack(t *testing.T) {
	os.Setenv("gateway_public_url", "http://localhost:8080")
	os.Setenv("gateway_pretty_url", "")

	event := &sdk.Event{
		Owner:   "alexellis",
		Service: "stack-deploy",
	}

	got := buildPublicStatusURL("success", sdk.BuildFunctionContext(event.Service), event)
	want := "http://localhost:8080/function/system-dashboard"

	if got != want {
		t.Errorf("building PublicURL: want %s, got %s", want, got)
		t.Fail()
	}
}

func TestBuildURLWithoutPrettyURL_WithOutSlash(t *testing.T) {
	os.Setenv("gateway_public_url", "http://localhost:8080")
	os.Setenv("gateway_pretty_url", "")

	event := &sdk.Event{
		Owner:   "alexellis",
		Service: "stack-deploy",
	}

	got := buildPublicStatusURL("success", sdk.BuildFunctionContext(event.Service), event)
	want := "http://localhost:8080/function/system-dashboard"

	if got != want {
		t.Errorf("building PublicURL: want %s, got %s", want, got)
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

	got := buildPublicStatusURL("success", sdk.BuildFunctionContext(event.Service), event)
	want := "https://alexellis.openfaas-cloud.com/tester"

	if got != want {
		t.Errorf("building PublicURL: want %s, got %s", want, got)
		t.Fail()
	}
}

func TestBuildURLWithUndefinedStatusGivesOriginalURL(t *testing.T) {

	event := &sdk.Event{
		Owner:   "alexellis",
		Service: "stack-deploy",
		URL:     "http://original-value.local",
	}

	got := buildPublicStatusURL("not-supported", sdk.BuildFunctionContext(event.Service), event)
	want := event.URL

	if got != want {
		t.Errorf("building PublicURL: want %s, got %s", want, got)
		t.Fail()
	}
}

func TestBuildURLWithFunctionAsPartOfSubDomain(t *testing.T) {
	os.Setenv("gateway_public_url", "http://localhost:8080")
	os.Setenv("gateway_pretty_url", "https://user.function.openfaas-cloud.com/function")

	event := &sdk.Event{
		Owner:   "alexellis",
		Service: "tester",
		URL:     "http://original-value.local",
	}

	got := buildPublicStatusURL("success", sdk.BuildFunctionContext(event.Service), event)
	want := "https://alexellis.function.openfaas-cloud.com/tester"

	if got != want {
		t.Errorf("building PublicURL: want %s, got %s", want, got)
		t.Fail()
	}
}

func TestBuildURLWithFunctionAsPartOfSubDomainHttp(t *testing.T) {
	os.Setenv("gateway_public_url", "http://localhost:8080")
	os.Setenv("gateway_pretty_url", "http://user.function.openfaas-cloud.com/function")

	event := &sdk.Event{
		Owner:   "alexellis",
		Service: "stack-deploy",
		URL:     "http://original-value.local",
	}

	got := buildPublicStatusURL("success", sdk.BuildFunctionContext(event.Service), event)
	want := "http://system.function.openfaas-cloud.com/dashboard/alexellis"

	if got != want {
		t.Errorf("building PublicURL: want %s, got %s", want, got)
		t.Fail()
	}
}

func TestBuildURLWithFunctionAsPartOfDomain(t *testing.T) {
	os.Setenv("gateway_public_url", "http://localhost:8080")
	os.Setenv("gateway_pretty_url", "https://user.functionopenfaas-cloud.com/function")

	event := &sdk.Event{
		Owner:   "alexellis",
		Service: "tester",
		URL:     "http://original-value.local",
	}

	got := buildPublicStatusURL("success", sdk.BuildFunctionContext(event.Service), event)
	want := "https://alexellis.functionopenfaas-cloud.com/tester"

	if got != want {
		t.Errorf("building PublicURL: want %s, got %s", want, got)
		t.Fail()
	}
}

func TestBuildURLWithFunctionAsPartOfDomainHttpStack(t *testing.T) {
	os.Setenv("gateway_public_url", "http://localhost:8080")
	os.Setenv("gateway_pretty_url", "http://user.functionopenfaas-cloud.com/function")

	event := &sdk.Event{
		Owner:   "alexellis",
		Service: "stack-deploy",
		URL:     "http://original-value.local",
	}
	want := "http://system.functionopenfaas-cloud.com/dashboard/alexellis"

	got := buildPublicStatusURL("success", sdk.BuildFunctionContext(event.Service), event)

	if got != want {
		t.Errorf("building PublicURL: want %s, got %s", want, got)
		t.Fail()
	}
}

func TestBuildURLFailsWithPrettyUrlStack(t *testing.T) {
	os.Setenv("gateway_public_url", "http://localhost:8080")
	os.Setenv("gateway_pretty_url", "https://user.functionopenfaas-cloud.com/function")

	event := &sdk.Event{
		Owner:   "alexellis",
		Service: "stack-deploy",
		URL:     "http://original-value.local",
	}
	want := "https://system.functionopenfaas-cloud.com/dashboard/alexellis"

	got := buildPublicStatusURL("failure", sdk.BuildFunctionContext(event.Service), event)

	if got != want {
		t.Errorf("building PublicURL: want %s, got %s", want, got)
		t.Fail()
	}
}

func TestBuildURLFailsWithPrettyUrl(t *testing.T) {
	os.Setenv("gateway_public_url", "http://localhost:8080")
	os.Setenv("gateway_pretty_url", "https://user.functionopenfaas-cloud.com/function")

	event := &sdk.Event{
		Owner:      "alexellis",
		Service:    "my-function-user-name-stuff",
		URL:        "http://original-value.local",
		Repository: "https://example.com/github.git",
		SHA:        "SomeSha",
	}
	want := "https://system.functionopenfaas-cloud.com/dashboard/alexellis/my-function-user-name-stuff/build-log?repoPath=alexellis/https://example.com/github.git&commitSHA=SomeSha"

	got := buildPublicStatusURL("failure", sdk.BuildFunctionContext(event.Service), event)

	if got != want {
		t.Errorf("building PublicURL: want %s, got %s", want, got)
		t.Fail()
	}
}

func TestBuildURLFailsWithPublicUrl(t *testing.T) {
	os.Setenv("gateway_public_url", "http://localhost:8080")
	os.Setenv("gateway_pretty_url", "")

	event := &sdk.Event{
		Owner:   "alexellis",
		Service: "stack-deploy",
		URL:     "http://original-value.local",
	}
	want := "http://localhost:8080/function/system-dashboard"

	got := buildPublicStatusURL("failure", sdk.BuildFunctionContext(event.Service), event)

	if got != want {
		t.Errorf("building PublicURL: want %s, got %s", want, got)
		t.Fail()
	}
}

func TestBuildPrettyURLNoUrl(t *testing.T) {
	gatewayPrettyUrl := ""

	want := ""

	event := &sdk.Event{
		Owner:   "alexellis",
		Service: "stack-deploy",
		URL:     "http://original-value.local",
	}

	got := buildPrettyURL(gatewayPrettyUrl, false, false, event)

	if got != want {
		t.Errorf("building url: want %s, got %s", want, got)
		t.Fail()
	}
}

func TestBuildPrettyURLFail(t *testing.T) {
	gatewayPrettyUrl := "https://user.functionopenfaas-cloud.com/function"

	want := "https://system.functionopenfaas-cloud.com/dashboard/Alexellis/some-random-service/build-log?repoPath=Alexellis/https://example.com/github.git&commitSHA=SomeSha"

	event := &sdk.Event{
		Owner:      "Alexellis",
		Service:    "some-random-service",
		Repository: "https://example.com/github.git",
		SHA:        "SomeSha",
	}

	got := buildPrettyURL(gatewayPrettyUrl, false, false, event)

	if got != want {
		t.Errorf("building url: want %s, got %s", want, got)
		t.Fail()
	}
}

func TestBuildPrettyURLSuccessStack(t *testing.T) {
	gatewayPrettyUrl := "https://user.functionopenfaas-cloud.com/function"

	want := "https://system.functionopenfaas-cloud.com/dashboard/Alexellis"

	event := &sdk.Event{
		Owner:   "Alexellis",
		Service: "test-function-service",
	}

	got := buildPrettyURL(gatewayPrettyUrl, true, true, event)

	if got != want {
		t.Errorf("building url: want %s, got %s", want, got)
		t.Fail()
	}
}

func TestBuildPrettyURLSuccess(t *testing.T) {
	gatewayPrettyUrl := "https://user.functionopenfaas-cloud.com/function"

	want := "https://alexellis.functionopenfaas-cloud.com/test-function-service"

	event := &sdk.Event{
		Owner:   "Alexellis",
		Service: "test-function-service",
	}

	got := buildPrettyURL(gatewayPrettyUrl, true, false, event)

	if got != want {
		t.Errorf("building url: want %s, got %s", want, got)
		t.Fail()
	}
}

func TestBuildPublicUrlSuccess(t *testing.T) {
	gatewayURL := "http://localhost:8080"

	want := "http://localhost:8080/function/alexellis-test-function-service"

	got := buildPublicURL(gatewayURL, "Alexellis", "test-function-service", true, false)

	if got != want {
		t.Errorf("building url: want %s, got %s", want, got)
		t.Fail()
	}
}

func TestBuildPublicUrlSuccessStack(t *testing.T) {
	gatewayURL := "http://localhost:8080"

	want := "http://localhost:8080/function/system-dashboard"

	got := buildPublicURL(gatewayURL, "Alexellis", "test-function-service", true, true)

	if got != want {
		t.Errorf("building url: want %s, got %s", want, got)
		t.Fail()
	}
}

func TestBuildPublicUrlFailStack(t *testing.T) {
	gatewayURL := "http://localhost:8080"

	want := "http://localhost:8080/function/system-dashboard"

	got := buildPublicURL(gatewayURL, "Alexellis", "test-function-service", false, true)

	if got != want {
		t.Errorf("building url: want %s, got %s", want, got)
		t.Fail()
	}
}

func TestBuildPublicUrlFail(t *testing.T) {
	gatewayURL := "http://localhost:8080"

	want := "http://localhost:8080/function/system-dashboard"

	got := buildPublicURL(gatewayURL, "Alexellis", "test-function-service", false, false)

	if got != want {
		t.Errorf("building url: want %s, got %s", want, got)
		t.Fail()
	}
}

func TestReplaceFunctionSuffixWithSlash(t *testing.T) {
	want := "https://function.function.function/dashboard"

	got := replaceFunctionSuffix("https://function.function.function/function/", "dashboard")
	if got != want {
		t.Errorf("building url: want %s, got %s", want, got)
		t.Fail()
	}
}

func TestReplaceFunctionSuffixNoSlash(t *testing.T) {
	want := "https://function.function.function/dashboard"

	got := replaceFunctionSuffix("https://function.function.function/function", "dashboard")
	if got != want {
		t.Errorf("building url: want %s, got %s", want, got)
		t.Fail()
	}
}

func TestReplaceFunctionSuffixNoSuffixFound(t *testing.T) {
	want := "https://function.function/dashboard/dashboard"

	got := replaceFunctionSuffix("https://function.function/dashboard", "dashboard")
	if got != want {
		t.Errorf("building url: want %s, got %s", want, got)
		t.Fail()
	}
}

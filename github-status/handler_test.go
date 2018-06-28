package function

import (
	"os"
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

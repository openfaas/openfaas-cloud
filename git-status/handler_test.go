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

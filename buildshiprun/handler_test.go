package function

import (
	"os"
	"testing"
)

func TestBuildURLWithoutPrettyURL_WithSlash(t *testing.T) {
	os.Setenv("gateway_public_url", "http://localhost:8080")
	os.Setenv("gateway_pretty_url", "")

	event := &eventInfo{
		owner:   "alexellis",
		service: "tester",
	}

	val := buildPublicStatusURL("success", event)
	want := "http://localhost:8080/function/alexellis-tester"

	if val != want {
		t.Errorf("building PublicURL: want %s, got %s", want, val)
		t.Fail()
	}
}

func TestBuildURLWithoutPrettyURL_WithOutSlash(t *testing.T) {
	os.Setenv("gateway_public_url", "http://localhost:8080")
	os.Setenv("gateway_pretty_url", "")

	event := &eventInfo{
		owner:   "alexellis",
		service: "tester",
	}

	val := buildPublicStatusURL("success", event)
	want := "http://localhost:8080/function/alexellis-tester"

	if val != want {
		t.Errorf("building PublicURL: want %s, got %s", want, val)
		t.Fail()
	}
}

func TestBuildURLWithPrettyURL(t *testing.T) {
	os.Setenv("gateway_public_url", "http://localhost:8080")
	os.Setenv("gateway_pretty_url", "https://user.openfaas-cloud.com/function")

	event := &eventInfo{
		owner:   "alexellis",
		service: "tester",
	}

	val := buildPublicStatusURL("success", event)
	want := "https://alexellis.openfaas-cloud.com/tester"

	if val != want {
		t.Errorf("building PublicURL: want %s, got %s", want, val)
		t.Fail()
	}
}

func TestBuildURLWithUndefinedStatusGivesOriginalURL(t *testing.T) {

	event := &eventInfo{
		owner:   "alexellis",
		service: "tester",
		url:     "http://original-value.local",
	}

	val := buildPublicStatusURL("not-supported", event)
	want := event.url

	if val != want {
		t.Errorf("building PublicURL: want %s, got %s", want, val)
		t.Fail()
	}
}

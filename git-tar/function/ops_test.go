package function

import (
	"fmt"
	"testing"

	// internal dependencies
	"github.com/openfaas/openfaas-cloud/sdk"
)

func genPushEvent(id int, cloneURL string, private bool) sdk.PushEvent {
	return sdk.PushEvent{
		Installation: sdk.PushEventInstallation{
			ID: id,
		},
		Repository: sdk.PushEventRepository{
			CloneURL: cloneURL,
			Private:  private,
		},
	}
}

var (
	defaultToken          = "authToken1234"
	defaultInstallationID = 12345
	defaultError          = fmt.Errorf("error message")
)

type authTokenStub struct {
	getTokenCalledTimes int
	returnError         bool
}

func (t *authTokenStub) getToken() (string, error) {
	t.getTokenCalledTimes++

	if t.returnError {
		return "", defaultError
	}

	return defaultToken, nil
}

func (t *authTokenStub) getInstallationID() int {
	return defaultInstallationID
}

func Test_getRepositoryURL_whenRepositoryIsNotPrivate(t *testing.T) {
	expected := "https://foo.bar/baz"
	pe := genPushEvent(12345, expected, false)
	at := &authTokenStub{}

	got, err := getRepositoryURL(pe, at)

	if err != nil {
		t.Error(err)
	}

	if got != expected {
		t.Errorf("Expected: %s, Got: %s", expected, got)
	}
}

func Test_getRepositoryURL_whenRepositoryIsPrivate(t *testing.T) {
	pe := genPushEvent(12345, "https://foo.bar/baz", true)
	at := &authTokenStub{}

	got, err := getRepositoryURL(pe, at)

	if err != nil {
		t.Error(err)
	}

	expected := fmt.Sprintf("https://%d:%s@foo.bar/baz", defaultInstallationID, defaultToken)

	if got != expected {
		t.Errorf("Expected: %s, Got: %s", expected, got)
	}

	if at.getTokenCalledTimes != 1 {
		t.Errorf("When repository is private githubAuthToken.getToken method should be called but it wasn't")
	}
}

func Test_getRepositoryURL_whenGetTokenReturnsError_WhenRepositoryIsPrivate(t *testing.T) {
	pe := genPushEvent(0, "", true)
	at := &authTokenStub{
		returnError: true,
	}

	expectedURL := ""

	gotURL, gotErr := getRepositoryURL(pe, at)

	if gotErr == nil {
		t.Errorf("Expected error")
	}

	if gotURL != expectedURL {
		t.Errorf("Expected URL: %s, Got: %s", expectedURL, gotURL)
	}
}

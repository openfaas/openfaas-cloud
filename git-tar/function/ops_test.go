package function

import (
	"fmt"
	"os"
	"strings"
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

func Test_formatTemplateRepos(t *testing.T) {
	formalRepos := []string{
		"https://github.com/openfaas/templates",
	}

	tests := []struct {
		title         string
		envRepos      string
		expectedRepos []string
		expectedError bool
	}{
		{
			title:         "Templates with no added custom repositories",
			envRepos:      "",
			expectedRepos: formalRepos,
			expectedError: false,
		},
		{
			title:         "Templates with single added custom repository",
			envRepos:      "https://github.com/my-custom/repo.git",
			expectedRepos: append(formalRepos, "https://github.com/my-custom/repo.git"),
			expectedError: false,
		},
		{
			title:         "Templates with two added custom repositories without spaces",
			envRepos:      "https://github.com/my-custom/repo.git,https://github.com/another/repo.git",
			expectedRepos: append(formalRepos, ([]string{"https://github.com/my-custom/repo.git", "https://github.com/another/repo.git"})...),
			expectedError: false,
		},
		{
			title:         "Variable set invalid",
			envRepos:      " ",
			expectedRepos: formalRepos,
			expectedError: true,
		},
		{
			title:         "Variable set with random symbols",
			envRepos:      "123randomzxc",
			expectedRepos: formalRepos,
			expectedError: true,
		},
		{
			title:         "Invalid github URLs (Missing `https://`)",
			envRepos:      "www.github.com/my-custom/repo.git",
			expectedRepos: formalRepos,
			expectedError: true,
		},
		{
			title:         "Setting values with spaces between commas",
			envRepos:      " , https://github.com/my-custom/repo.git, https://github.com/another/repo.git, ",
			expectedRepos: append(formalRepos, ([]string{"https://github.com/my-custom/repo.git", "https://github.com/another/repo.git"})...),
			expectedError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			os.Setenv("custom_templates", test.envRepos)
			templateRepos, err := formatTemplateRepos()
			if err != nil && !test.expectedError {
				t.Errorf("want: no error, got: %t", err)
			}
			if err == nil && test.expectedError {
				t.Errorf("want: %t got: nil", err)
			}
			if len(templateRepos) != len(test.expectedRepos) {
				t.Errorf("want: \n`%s` \ngot: \n`%s`",
					strings.Join(test.expectedRepos, " "),
					strings.Join(templateRepos, " "))
			}
			for i := 0; i < len(test.expectedRepos); i++ {
				if test.expectedRepos[i] != templateRepos[i] {
					t.Errorf("want: %s got: %s", test.expectedRepos[i], templateRepos[i])
				}
			}
		})
	}
}

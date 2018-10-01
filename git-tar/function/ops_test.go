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

func Test_formatTemplateReposValid(t *testing.T) {
	formalRepos := []string{
		"https://github.com/openfaas/templates",
		"https://github.com/openfaas-incubator/node8-express-template.git",
		"https://github.com/openfaas-incubator/golang-http-template.git",
	}

	tests := []struct {
		title         string
		envRepos      string
		expectedRepos []string
	}{
		{
			title:         "Templates with no added custom repositories",
			envRepos:      "",
			expectedRepos: formalRepos,
		},
		{
			title:         "Templates with single added custom repository",
			envRepos:      "https://github.com/my-custom/repo.git",
			expectedRepos: append(formalRepos, "https://github.com/my-custom/repo.git"),
		},
		{
			title:         "Templates with two added custom repositories without spaces",
			envRepos:      "https://github.com/my-custom/repo.git,https://github.com/another/repo.git",
			expectedRepos: append(formalRepos, ([]string{"https://github.com/my-custom/repo.git", "https://github.com/another/repo.git"})...),
		},
	}
	for _, test := range tests {
		os.Setenv("custom_templates", test.envRepos)
		t.Run(test.title, func(t *testing.T) {
			templateRepos := formatTemplateRepos()
			for _, templateRepo := range templateRepos {
				for final, expectedRepo := range test.expectedRepos {
					if expectedRepo == templateRepo {
						continue
					}
					if final == len(test.expectedRepos) {
						t.Errorf("Expecting repositories: \n`%s` \ngot: \n`%s`",
							strings.Join(test.expectedRepos, " "),
							strings.Join(templateRepos, " "))
					}

				}
			}
		})
	}
}

func Test_formatTemplateReposUnvalid(t *testing.T) {
	formalRepos := []string{
		"https://github.com/openfaas/templates",
		"https://github.com/openfaas-incubator/node8-express-template.git",
		"https://github.com/openfaas-incubator/golang-http-template.git",
	}

	tests := []struct {
		title         string
		envRepos      string
		expectedRepos []string
	}{
		{
			title:         "Variable set invalid",
			envRepos:      " ",
			expectedRepos: formalRepos,
		},
		{
			title:         "Variable set with random symbols",
			envRepos:      "123randomzxc",
			expectedRepos: formalRepos,
		},
		{
			title:         "Invalid github URLs (Missing `https://`)",
			envRepos:      "www.github.com/my-custom/repo.git",
			expectedRepos: formalRepos,
		},
		{
			title:         "Setting values with spaces between commas",
			envRepos:      " , https://github.com/my-custom/repo.git, https://github.com/another/repo.git, ",
			expectedRepos: append(formalRepos, ([]string{"https://github.com/my-custom/repo.git", "https://github.com/another/repo.git"})...),
		},
	}
	for _, test := range tests {
		os.Setenv("custom_templates", test.envRepos)
		t.Run(test.title, func(t *testing.T) {
			templateRepos := formatTemplateRepos()
			for _, templateRepo := range templateRepos {
				for final, expectedRepo := range test.expectedRepos {
					if expectedRepo == templateRepo {
						continue
					}
					if final == len(test.expectedRepos) {
						t.Errorf("Expecting repositories: \n`%s` \ngot: \n`%s`",
							strings.Join(test.expectedRepos, " "),
							strings.Join(templateRepos, " "))
					}

				}
			}
		})
	}
}

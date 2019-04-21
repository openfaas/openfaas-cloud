package function

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	// internal dependencies
	"github.com/openfaas/faas-cli/stack"
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
		"https://github.com/openfaas-incubator/node10-express-template.git",
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
		"https://github.com/openfaas-incubator/node10-express-template.git",
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

func Test_formatGitLabCloneURL(t *testing.T) {
	tests := []struct {
		title            string
		event            sdk.PushEvent
		token            string
		expectedCloneURL string
	}{
		{
			title: "We have the fields populated right",
			event: sdk.PushEvent{
				Repository: sdk.PushEventRepository{
					Owner:    sdk.Owner{Login: "martindekov"},
					CloneURL: "https://gitlab.example.io/martindekov/gitlab-playground.git",
				},
			},
			token:            "zxcvasd123",
			expectedCloneURL: "https://martindekov:zxcvasd123@gitlab.example.io/martindekov/gitlab-playground.git",
		},
		{
			title:            "We have the struct not populated right but token exists",
			event:            sdk.PushEvent{},
			token:            "zxcvasd123",
			expectedCloneURL: "https://:zxcvasd123@",
		},
		{
			title:            "We have nothing populated right and token does not exist",
			event:            sdk.PushEvent{},
			token:            "",
			expectedCloneURL: "https://:@",
		},
	}
	var expectedError error
	for _, test := range tests {
		cloneURL, formatErr := formatGitLabCloneURL(test.event, test.token)
		t.Run("Properly formatted URL", func(t *testing.T) {
			if cloneURL != test.expectedCloneURL {
				t.Errorf("Expected URL: %s got: %s", test.expectedCloneURL, cloneURL)
			}
			if formatErr != expectedError {
				t.Errorf("Expected error: %v got: %v", expectedError, formatErr)
			}
		})
	}
}

func Test_existingTemplates(t *testing.T) {
	expectingTemplates := []string{"go", "python", "rust"}
	templateDir, dirErr := mockTempTemplatesDir(expectingTemplates, "template")
	if dirErr != nil {
		t.Errorf("Error while mocking template dir: %s", dirErr)
	}
	defer os.RemoveAll(templateDir)
	dir := os.TempDir()
	templates, templatesError := existingTemplates(dir)
	if templatesError != nil {
		t.Errorf("Error while checking templates: %s", templatesError)
	}
	for _, template := range templates {
		for lastTemplate, expectedTemplate := range expectingTemplates {
			if template == expectedTemplate {
				break
			}
			if lastTemplate == len(expectingTemplates)-1 &&
				expectedTemplate != template {
				t.Errorf("Error template: %s not found in: %v", template, expectingTemplates)
			}
		}
	}
}

func Test_checkCompatibleTemplates(t *testing.T) {
	tests := []struct {
		title              string
		functionDefinition *stack.Services
		expectedError      error
	}{
		{
			title: "Function language exists in the fetched templates",
			functionDefinition: &stack.Services{Functions: map[string]stack.Function{
				"fn1": stack.Function{Language: "go"},
			},
			},
			expectedError: nil,
		},
		{
			title: "Function language does not exist in the fetched templates",
			functionDefinition: &stack.Services{Functions: map[string]stack.Function{
				"fn1": stack.Function{Language: "smalltalk"},
			},
			},
			expectedError: fmt.Errorf("Not supported language: `smalltalk` for function: `fn1`"),
		},
	}
	existingTemplates := []string{"go", "python", "rust", "java"}
	templateDir, dirErr := mockTempTemplatesDir(existingTemplates, "template")
	if dirErr != nil {
		t.Errorf("Error while createding template dir: %s", dirErr)
	}
	defer os.RemoveAll(templateDir)
	tmpDir := os.TempDir()
	for _, test := range tests {
		compatibilityError := checkCompatibleTemplates(test.functionDefinition, tmpDir)
		if compatibilityError != nil && compatibilityError != test.expectedError {
			if compatibilityError.Error() != test.expectedError.Error() {
				t.Errorf("Expected error: `%s`, got: `%s`",
					test.expectedError.Error(),
					compatibilityError.Error())
			}
		}
	}
}

func mockTempTemplatesDir(files []string, directory string) (string, error) {
	permissions := 0744
	tmpDir := os.TempDir()
	templatesDir := fmt.Sprintf("%s/%s", tmpDir, directory)
	dirErr := os.MkdirAll(templatesDir, os.FileMode(permissions))
	if dirErr != nil {
		return templatesDir, dirErr
	}
	for _, file := range files {
		dirErr := os.MkdirAll(templatesDir+"/"+file, os.FileMode(permissions))
		if dirErr != nil {
			return templatesDir, dirErr
		}
	}
	fileErr := ioutil.WriteFile(templatesDir+"/some.txt", []byte{}, os.FileMode(permissions))
	if dirErr != nil {
		return templatesDir, fileErr
	}
	return templatesDir, nil
}

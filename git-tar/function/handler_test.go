package function

import (
	"testing"

	"github.com/openfaas/faas-cli/stack"
)

func Test_getRawURL(t *testing.T) {
	var pushEvents = []struct {
		RepositoryURL        string
		RepositoryOwnerLogin string
		RepositoryName       string
		AfterCommitID        string
		SCM                  string
		Expected             string
	}{
		{
			RepositoryURL:        "https://gitlab.url.io/myuser/myrepo",
			RepositoryOwnerLogin: "myuser",
			RepositoryName:       "myrepo",
			SCM:                  "github",
			Expected:             "https://raw.githubusercontent.com/myuser/myrepo/master/stack.yml",
		},
		{
			RepositoryURL:        "https://gitlab.url.io/myuser/myrepo",
			RepositoryOwnerLogin: "myuser",
			RepositoryName:       "myrepo",
			SCM:                  "github",
			Expected:             "https://raw.githubusercontent.com/myuser/myrepo/master/stack.yml",
		},
		{
			RepositoryURL:        "https://gitlab.url.io/myuser/myrepo",
			RepositoryOwnerLogin: "myuser",
			RepositoryName:       "myrepo",
			SCM:                  "gitlab",
			Expected:             "https://gitlab.url.io/myuser/myrepo/raw/master/stack.yml",
		},
		{
			RepositoryURL:        "https://gitlab.url.io/myuser/myrepo",
			RepositoryOwnerLogin: "myuser",
			RepositoryName:       "myrepo",
			SCM:                  "gitlab",
			Expected:             "https://gitlab.url.io/myuser/myrepo/raw/master/stack.yml",
		},
	}

	for _, pushEvent := range pushEvents {
		addr, _ := getRawURL(pushEvent.SCM, pushEvent.RepositoryURL, pushEvent.RepositoryOwnerLogin, pushEvent.RepositoryName)
		if addr != pushEvent.Expected {
			t.Errorf("Want \"%s\", got \"%s\"", pushEvent.Expected, addr)
		}
	}

}

func Test_hasDockerfileFunction(t *testing.T) {
	var cases = []struct {
		title    string
		input    map[string]stack.Function
		expected bool
	}{
		{
			title: "no dockerfile function",
			input: map[string]stack.Function{
				"test": stack.Function{
					Language: "go",
				},
			},
			expected: false,
		},
		{
			title: "with a dockerfile function",
			input: map[string]stack.Function{
				"test": stack.Function{
					Language: "Dockerfile",
				},
			},
			expected: true,
		},
	}

	for _, c := range cases {
		t.Run(c.title, func(t *testing.T) {
			result := hasDockerfileFunction(c.input)
			if result != c.expected {
				t.Errorf("Expected %v but got %v instead", c.expected, result)
			}
		})
	}
}

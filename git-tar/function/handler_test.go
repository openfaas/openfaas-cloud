package function

import (
	"testing"
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

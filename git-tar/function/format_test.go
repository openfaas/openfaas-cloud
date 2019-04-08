package function

import (
	"net/url"
	"os"
	"testing"

	"github.com/openfaas/faas-cli/stack"
)

func Test_createCloneURL(t *testing.T) {

	testURL := "https://github.com/alexellis/derek"

	url1, _ := url.Parse(testURL)
	url1.User = url.UserPassword("alex", "test1234")

	want := "https://alex:test1234@github.com/alexellis/derek"

	if url1.String() != want {
		t.Errorf("clone URL want %s, got %s", want, url1.String())
		t.Fail()
	}

}

func Test_FormatImageShaTag_PrivateRepo_WithTag_NoStackPrefix(t *testing.T) {
	function := &stack.Function{
		Image: "func:0.2",
	}

	owner := "alexellis"
	repo := "go-fns-tester"
	sha := "04b8e44988"
	os.Setenv("build_branch", "master")

	name := formatImageShaTag("registry:5000", function, sha, owner, repo)

	want := "registry:5000/" + owner + "/" + repo + "-func:0.2-master-04b8e44"
	if name != want {
		t.Errorf("Want \"%s\", got \"%s\"", want, name)
	}
}

func Test_FormatImageShaTag_PrivateRepo_WithTag(t *testing.T) {
	function := &stack.Function{
		Image: "alexellis2/func:0.2",
	}

	owner := "alexellis"
	repo := "go-fns-tester"
	sha := "04b8e44988"
	os.Setenv("build_branch", "master")

	name := formatImageShaTag("registry:5000", function, sha, owner, repo)

	want := "registry:5000/" + owner + "/" + repo + "-func:0.2-master-04b8e44"
	if name != want {
		t.Errorf("Want \"%s\", got \"%s\"", want, name)
	}
}

func Test_FormatImageShaTag_PrivateRepo_NoTag(t *testing.T) {
	function := &stack.Function{
		Image: "alexellis2/func",
	}

	owner := "alexellis"
	repo := "go-fns-tester"
	sha := "04b8e44988"
	os.Setenv("build_branch", "master")

	name := formatImageShaTag("registry:5000", function, sha, owner, repo)

	want := "registry:5000/" + owner + "/" + repo + "-func:latest-master-04b8e44"
	if name != want {
		t.Errorf("Want \"%s\", got \"%s\"", want, name)
	}
}

func Test_FormatImageShaTag_SharedRepo_NoTag(t *testing.T) {
	function := &stack.Function{
		Image: "alexellis2/func",
	}

	owner := "alexellis"
	repo := "go-fns-tester"
	sha := "04b8e44988"
	os.Setenv("build_branch", "master")
	name := formatImageShaTag("docker.io/of-community/", function, sha, owner, repo)

	want := "docker.io/of-community/" + owner + "-" + repo + "-func:latest-master-04b8e44"
	if name != want {
		t.Errorf("Want \"%s\", got \"%s\"", want, name)
	}
}

package function

import (
	"testing"

	"github.com/openfaas/openfaas-cloud/sdk"
)

func Test_getPath(t *testing.T) {
	got := getPath("pipeline", &sdk.PipelineLog{
		RepoPath:  "alexellis/super-cake",
		CommitSHA: "af6db",
		Function:  "slack-fn",
	})
	want := "pipeline/alexellis/super-cake/af6db/slack-fn/build.log"
	if got != want {
		t.Errorf("got: %s, but want: %s", got, want)
		t.Fail()
	}
}

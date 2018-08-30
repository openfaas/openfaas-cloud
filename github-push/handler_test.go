package function

import (
	"os"
	"testing"

	"github.com/openfaas/openfaas-cloud/sdk"
)

func Test_Handle_EmptyEvent(t *testing.T) {
	audit = sdk.NilLogger{}
	os.Setenv("Http_X_Github_Event", "")

	res := Handle([]byte{})
	want := "github-push cannot handle event: "
	if res != want {
		t.Errorf("want error: \"%s\", got: \"%s\"", want, res)
		t.Fail()
	}
}

func Test_Handle_IssueComment(t *testing.T) {
	audit = sdk.NilLogger{}
	os.Setenv("Http_X_Github_Event", "IssueComment")

	res := Handle([]byte{})
	want := "github-push cannot handle event: IssueComment"
	if res != want {
		t.Errorf("want error: \"%s\", got: \"%s\"", want, res)
		t.Fail()
	}
}

package function

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/openfaas/openfaas-cloud/sdk"
)

type HTTPHandler struct {
}

func (h HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("alexellis\n"))
}

func Test_Handle_Push_InvalidBranch(t *testing.T) {
	audit = sdk.NilLogger{}
	os.Setenv("Http_X_Github_Event", "push")
	os.Setenv("validate_hmac", "false")
	os.Setenv("validate_customers", "false")

	res := Handle([]byte(
		`{"ref":"refs/heads/staging"}`,
	))

	want := "refusing to build target branch: refs/heads/staging, want branch: master"
	if res != want {
		t.Errorf("want error: \"%s\", got: \"%s\"", want, res)
		t.Fail()
	}
}

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

func Test_Handle_ValidateCustomers_Matched(t *testing.T) {
	server := httptest.NewServer(&HTTPHandler{})
	os.Setenv("Http_X_Github_Event", "push")
	os.Setenv("validate_customers", "true")
	os.Setenv("customers_url", server.URL)
	res := Handle([]byte(
		`{"ref":"refs/heads/master","repository":{ "owner": { "login": "alexellis" } }}`,
	))
	// This error is as far as we can get right now without subbing more code.
	secretErr := "unable to read secret"
	if !strings.Contains(res, secretErr) {
		t.Errorf("want error: \"%s\", got: \"%s\"", secretErr, res)
		t.Fail()
	}
}

package function

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/openfaas/openfaas-cloud/sdk"
)

type HTTPHandler struct {
}

func (h HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("alexellis\n"))
}

func Test_validCustomer_Found_Passes(t *testing.T) {
	res := validCustomer([]string{"rgee0", "alexellis"}, "alexellis")

	want := true
	if res != want {
		t.Errorf("want error: \"%t\", got: \"%t\"", want, res)
		t.Fail()
	}
}

func Test_validCustomer_NotFound_Fails(t *testing.T) {
	res := validCustomer([]string{"alexellis"}, "rgee0")

	want := false
	if res != want {
		t.Errorf("want error: \"%t\", got: \"%t\"", want, res)
		t.Fail()
	}
}

func Test_validCustomer_EmptyInput_Fails(t *testing.T) {
	res := validCustomer([]string{"alexellis"}, "")

	want := false

	if res != want {
		t.Errorf("want error: \"%t\", got: \"%t\"", want, res)
		t.Fail()
	}
}

func Test_Handle_ValidateCustomersInvalid(t *testing.T) {
	audit = sdk.NilLogger{}
	server := httptest.NewServer(&HTTPHandler{})
	defer server.Close()

	os.Setenv("Http_X_Github_Event", "push")
	os.Setenv("validate_customers", "true")
	os.Setenv("customers_url", server.URL)

	res := Handle([]byte(
		`{"ref":"refs/heads/master","repository":{ "owner": { "login": "rgee0" } }}`,
	))

	want := "Customer: rgee0 not found in CUSTOMERS file via " + server.URL
	if res != want {
		t.Errorf("want error: \"%s\", got: \"%s\"", want, res)
		t.Fail()
	}
}

func Test_Handle_ValidateCustomers_Matched(t *testing.T) {
	audit = sdk.NilLogger{}
	server := httptest.NewServer(&HTTPHandler{})

	os.Setenv("Http_X_Github_Event", "push")
	os.Setenv("validate_customers", "true")
	os.Setenv("customers_url", server.URL)

	res := Handle([]byte(
		`{"ref":"refs/heads/master","repository":{ "owner": { "login": "alexellis" } }}`,
	))

	// This error is as far as we can get right now without subbing more code.
	want := "unable to read secret: /var/openfaas/secrets/payload-secret, error: open /var/openfaas/secrets/payload-secret: no such file or directory"
	if res != want {
		t.Errorf("want error: \"%s\", got: \"%s\"", want, res)
		t.Fail()
	}
}

func Test_Handle_Push_InvalidBranch(t *testing.T) {
	audit = sdk.NilLogger{}
	os.Setenv("Http_X_Github_Event", "push")
	os.Setenv("validate_customers", "false")

	res := Handle([]byte(
		`{"ref":"refs/heads/staging"}`,
	))

	want := "refusing to build non-master branch: refs/heads/staging"
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

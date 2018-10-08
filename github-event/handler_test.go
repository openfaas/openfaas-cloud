package function

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
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
	server := httptest.NewServer(&HTTPHandler{})
	defer server.Close()

	os.Setenv("Http_X_Github_Event", "push")
	os.Setenv("validate_customers", "true")
	os.Setenv("customers_url", server.URL)

	res := Handle([]byte(
		`{"sender" : { "login" : "rgee0" }}`,
	))

	want := "Customer: rgee0 not found in CUSTOMERS file via " + server.URL
	if res != want {
		t.Errorf("want error: \"%s\", got: \"%s\"", want, res)
		t.Fail()
	}
}

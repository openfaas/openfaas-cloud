package function

import (
	"encoding/json"
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

func Test_validateCustomers_UserNotFound(t *testing.T) {
	os.Unsetenv("Http_Query")

	s := httptest.NewServer(HTTPHandler{})

	customer := sdk.Customer{
		Sender: sdk.Sender{
			Login: "mark",
		},
	}

	res, _ := json.Marshal(customer)
	err := validateCustomers(res, s.URL)

	if err == nil {
		t.Errorf("Expected sender to be invalid and to generate an error")
	}
}

func Test_validateCustomers_UserFound(t *testing.T) {
	os.Unsetenv("Http_Query")

	s := httptest.NewServer(HTTPHandler{})

	customer := sdk.Customer{
		Sender: sdk.Sender{
			Login: "alexellis",
		},
	}

	res, _ := json.Marshal(customer)
	err := validateCustomers(res, s.URL)

	if err != nil {
		t.Errorf("Expected sender to be valid, but got error: %s", err.Error())
	}
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
	os.Unsetenv("Http_Query")

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

func Test_Handle_Event(t *testing.T) {
	os.Unsetenv("Http_Query")

	audit = sdk.NilLogger{}
	os.Setenv("Http_X_Hub_Signature", "")
	var events = []struct {
		scenario          string
		header            string
		action            string
		validateCustomers string
		validateHmac      string
		want              string
	}{
		{
			scenario:          "Empty event",
			header:            "",
			action:            "",
			validateCustomers: "false",
			validateHmac:      "false",
			want:              "github-event cannot handle event: ",
		},
		{
			scenario:          "Non-supported event",
			header:            "other",
			action:            "",
			validateCustomers: "false",
			validateHmac:      "false",
			want:              "github-event cannot handle event: other",
		},
		{
			scenario:          "Validate customers",
			header:            "push",
			action:            "",
			validateCustomers: "true",
			validateHmac:      "false",
			want:              "Customer:  not found in CUSTOMERS file via ",
		},
		{
			scenario:          "Push event",
			header:            "push",
			action:            "",
			validateCustomers: "false",
			validateHmac:      "false",
			want:              "unable to read secret: /var/openfaas/secrets/payload-secret, error: open /var/openfaas/secrets/payload-secret: no such file or directory",
		},
	}
	for _, event := range events {
		t.Run(event.scenario, func(t *testing.T) {
			os.Setenv("Http_X_Github_Event", event.header)
			os.Setenv("validate_customers", event.validateCustomers)

			req := []byte{}
			audit = sdk.NilLogger{}
			server := httptest.NewServer(&HTTPHandler{})

			if event.validateCustomers == "true" {
				os.Setenv("customers_url", server.URL)
				req = []byte(
					`{"ref":"refs/heads/master","repository":{ "owner": { "login": "alexellis" } }}`,
				)
				event.want = event.want + server.URL
			}

			res := Handle(req)

			if res != event.want {
				t.Errorf("want %s, but got %s", event.want, res)
			}
		})
	}
}

func Test_RedirectSetupAction(t *testing.T) {
	os.Setenv("Http_Query", "setup_action=install")
	got := Handle([]byte(""))
	want := "Installation completed, please return to the installation guide."
	if got != want {
		t.Errorf("want %s, got %s", want, got)
	}

}

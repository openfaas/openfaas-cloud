package sdk

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func Test_ValidateCustomers(t *testing.T) {
	tests := []struct {
		title        string
		value        string
		expectedBool bool
	}{
		{
			title:        "environmental variable `validate_customers` is unset",
			value:        "",
			expectedBool: true,
		},
		{
			title:        "environmental variable `validate_customers` is set to true",
			value:        "true",
			expectedBool: true,
		},
		{
			title:        "environmental variable `validate_customers` is set to 1",
			value:        "1",
			expectedBool: true,
		},
		{
			title:        "environmental variable `validate_customers` is set with random value",
			value:        "random",
			expectedBool: true,
		},
		{
			title:        "environmental variable `validate_customers` is set with explicit `0`",
			value:        "0",
			expectedBool: false,
		},
		{
			title:        "environmental variable `validate_customers` is set with explicit `false`",
			value:        "false",
			expectedBool: false,
		},
	}
	customersEnvVar := "validate_customers"
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			os.Setenv(customersEnvVar, test.value)
			value := ValidateCustomers()
			if value != test.expectedBool {
				t.Errorf("Expected value: %v got: %v", test.expectedBool, value)
			}
		})
	}
}

func Test_ValidateCustomerList(t *testing.T) {
	tests := []struct {
		Title        string
		CustomerList []string
		Output       bool
	}{
		{
			Title:        "Customer names without - ",
			CustomerList: []string{"alexellis", "stefan", "ivana"},
			Output:       true,
		},
		{
			Title:        "Customer names with -",
			CustomerList: []string{"alexellis", "alexellis-suffix", "stefan", "ivana"},
			Output:       false,
		},
		{
			Title:        "Customer names with -",
			CustomerList: []string{"alexellis", "stefan-prodan", "stefan", "ivana"},
			Output:       false,
		},
		{
			Title:        "Customer names with -",
			CustomerList: []string{"alexellis", "stefan-prodan", "ivana"},
			Output:       true,
		},
		{
			Title:        "Customer names without -",
			CustomerList: []string{"alexellis", "alexe", "stefan", "ivana"},
			Output:       true,
		},
	}

	for _, test := range tests {
		value := ValidateCustomerList(test.CustomerList)
		if value != test.Output {
			t.Errorf("Expected value: %v, got: %v", test.Output, value)
		}
	}
}

func TestGet_InvalidCustomerLiveGitHubFile(t *testing.T) {
	c := NewCustomers("", "")

	valid := []string{"not-alexellis"}

	for _, user := range valid {
		val, err := c.Get(user)
		if err != nil {
			t.Errorf("error fetching users: %s", err.Error())
			t.Fail()
		}

		if val != false {
			t.Errorf("user %s should not be a customer, but was", user)
			t.Fail()
		}
	}
}

func TestGet_ExistingCustomerLiveGitHubFile(t *testing.T) {
	c := NewCustomers("", "")

	valid := []string{"alexellis", "rgee0", "LucasRoesler"}

	for _, user := range valid {
		val, err := c.Get(user)
		if err != nil {
			t.Errorf("error fetching users: %s", err.Error())
			t.Fail()
		}

		if val != true {
			t.Errorf("user %s should be a customer, but wasn't", user)
			t.Fail()
		}
	}
}

func TestGet_FromFile(t *testing.T) {

	tmpPath := path.Join(os.TempDir(), "customers")
	writeErr := ioutil.WriteFile(tmpPath, []byte(`openfaas
inlets`), 0700)
	if writeErr != nil {
		t.Error(writeErr)
		t.Fail()
	}

	defer func() {
		os.RemoveAll(tmpPath)
	}()

	c := NewCustomers(tmpPath, "")

	valid := []string{"openfaas", "inlets"}

	for _, user := range valid {
		val, err := c.Get(user)
		if err != nil {
			t.Errorf("error fetching users: %s", err.Error())
			t.Fail()
		}

		if val != true {
			t.Errorf("user %s should be a customer, but wasn't", user)
			t.Fail()
		}
	}
}

func TestformatUsername_TrimsNewline(t *testing.T) {
	want := `alexellis`
	got := formatUsername(`alexelllis
`)
	if got != want {
		t.Errorf(`want %q, got %q`, want, got)
	}
}

func TestformatUsername_TrimsReturn(t *testing.T) {
	want := `alexellis`
	got := formatUsername("alexelllis\r")
	if got != want {
		t.Errorf(`want %q, got %q`, want, got)
	}
}

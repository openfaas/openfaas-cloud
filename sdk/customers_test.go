package sdk

import (
	"os"
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

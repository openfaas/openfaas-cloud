package sdk

import "os"

// ValidateCustomers checks environmental
// variable validate_customers if customer
// validation is explicitly disabled
func ValidateCustomers() bool {
	if val, exists := os.LookupEnv("validate_customers"); exists {
		return val != "false" && val != "0"
	}
	return true
}

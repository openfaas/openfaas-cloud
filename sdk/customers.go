package sdk

import (
	"os"
	"strings"
)

// ValidateCustomers checks environmental
// variable validate_customers if customer
// validation is explicitly disabled
func ValidateCustomers() bool {
	if val, exists := os.LookupEnv("validate_customers"); exists {
		return val != "false" && val != "0"
	}
	return true
}

//ValidateCustomerList validate customer names list
func ValidateCustomerList(customers []string) bool {
	for i, customerName := range customers {
		for j, cn := range customers {

			if i != j {
				if strings.HasPrefix(cn, customerName+"-") {
					return false
				}
			}
		}
	}

	return true
}

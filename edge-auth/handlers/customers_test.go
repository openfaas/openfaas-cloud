package handlers

import (
	"testing"
)

func TestGet_InvalidCustomerLiveGitHubFile(t *testing.T) {
	c := NewCustomers()

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
	c := NewCustomers()

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

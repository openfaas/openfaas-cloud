package handlers

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/openfaas/openfaas-cloud/sdk"
	"sync"
	"testing"
	"time"
)

func Test_isInOrganisations_NoOrgs(t *testing.T) {
	want := false
	claims := OpenFaaSCloudClaims{
		Name:           "claim",
		AccessToken:    "token",
		Organizations:  "",
		StandardClaims: jwt.StandardClaims{},
	}

	usernames := make(map[string]string)
	usernames["user"] = "user"
	customers := sdk.Customers{
		Usernames:     &usernames,
		Sync:          &sync.Mutex{},
		Expires:       time.Now().Add(100 * time.Second),
		CustomersURL:  "",
		CustomersPath: "",
	}
	got := isInOrganisations(claims, &customers)

	if want != got {
		t.Error("didn't expect to find user's org in Customers but did")
		t.Fail()
	}
}

func Test_isInOrganisations_OrgsNotMember(t *testing.T) {
	want := false
	claims := OpenFaaSCloudClaims{
		Name:           "claim",
		AccessToken:    "token",
		Organizations:  "this,that",
		StandardClaims: jwt.StandardClaims{},
	}

	usernames := make(map[string]string)
	usernames["user"] = "user"
	customers := sdk.Customers{
		Usernames:     &usernames,
		Sync:          &sync.Mutex{},
		Expires:       time.Now().Add(100 * time.Second),
		CustomersURL:  "",
		CustomersPath: "",
	}
	got := isInOrganisations(claims, &customers)

	if want != got {
		t.Error("didn't expect to find user's org in Customers but did")
		t.Fail()
	}
}

func Test_isInOrganisations_IsInOrgs(t *testing.T) {
	want := true
	claims := OpenFaaSCloudClaims{
		Name:           "claim",
		AccessToken:    "token",
		Organizations:  "this,that",
		StandardClaims: jwt.StandardClaims{},
	}

	usernames := make(map[string]string)
	usernames["this"] = "this"
	customers := sdk.Customers{
		Usernames:     &usernames,
		Sync:          &sync.Mutex{},
		Expires:       time.Now().Add(100 * time.Second),
		CustomersURL:  "",
		CustomersPath: "",
	}
	got := isInOrganisations(claims, &customers)

	if want != got {
		t.Error("wanted to find user's org in Customers but didn't")
		t.Fail()
	}
}

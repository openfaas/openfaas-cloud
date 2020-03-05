package provider

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestSupportedProviders_IsSupported_whenProviderIsSupported(t *testing.T) {
	if !IsSupported("github") {
		t.Errorf("Want: \"%v\". Got: \"%v\"", true, false)
	}

	if !IsSupported("gitHub") {
		t.Errorf("Want: \"%v\". Got: \"%v\"", true, false)
	}

	if !IsSupported("GitLab") {
		t.Errorf("Want: \"%v\". Got: \"%v\"", true, false)
	}
}

func TestSupportedProviders_IsSupported_whenProviderIsNotSupported(t *testing.T) {
	if IsSupported("foobar.com") {
		t.Errorf("Want: \"%v\". Got: \"%v\"", false, true)
	}
}

func TestUnmarshalOrg(t *testing.T) {
	wantID := 49474643
	wantLogin := "teamserverless"
	orgVal := []byte(fmt.Sprintf(`{"login": "%s","id": %d}`, wantLogin, wantID))

	org := Organization{}
	err := json.Unmarshal(orgVal, &org)

	if err != nil {
		t.Errorf("received error, wanted none: %s", err.Error())
		t.Fail()
	}
	if org.ID != wantID {
		t.Errorf("ID want: %d, got: %d", wantID, org.ID)
		t.Fail()
	}
	if org.Login != wantLogin {
		t.Errorf("Login want: %s, got: %s", wantLogin, org.Login)
		t.Fail()
	}
}

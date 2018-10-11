package provider

import "testing"

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
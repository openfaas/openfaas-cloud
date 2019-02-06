package handlers

import (
	"fmt"
	"net/url"
	"testing"
)

func TestCombineURL_BuildsValidURL(t *testing.T) {
	want := "https://www.google.com/query/this"
	got := combineURL("https://www.google.com/query/", "/this")

	if want != got {
		t.Errorf("combineURL want: %s, got: %s", want, got)
		t.Fail()
	}
}

func Test_buildGitLabURL(t *testing.T) {
	c := &Config{
		OAuthProviderBaseURL:   "https://foo.bar",
		ClientID:               "baz",
		ExternalRedirectDomain: "http://bazfoz.com",
	}

	expectedURL, _ := url.Parse(fmt.Sprintf(
		"https://foo.bar/oauth/authorize?response_type=code&client_id=%s&redirect_uri=%s",
		c.ClientID,
		fmt.Sprintf("%s/oauth2/authorized", c.ExternalRedirectDomain),
	))
	expectedQuery := expectedURL.Query()

	gotURL := buildGitLabURL(c)
	gotQuery := gotURL.Query()

	if expectedURL.Host != gotURL.Host {
		t.Errorf("Expected host: \"%s\". Got: \"%s\"", expectedURL.Host, gotURL.Host)
	}

	if expectedURL.Path != gotURL.Path {
		t.Errorf("Expected path: \"%s\". Got: \"%s\"", expectedURL.Path, gotURL.Path)
	}

	if expectedQuery.Get("response_type") != gotQuery.Get("response_type") {
		t.Errorf(
			"Expected query.response_type: \"%s\". Got: \"%s\"",
			expectedQuery.Get("response_type"),
			gotQuery.Get("response_type"),
		)
	}

	if expectedQuery.Get("client_id") != gotQuery.Get("client_id") {
		t.Errorf(
			"Expected query.client_id: \"%s\". Got: \"%s\"",
			expectedQuery.Get("client_id"),
			gotQuery.Get("client_id"),
		)
	}

	if expectedQuery.Get("redirect_uri") != gotQuery.Get("redirect_uri") {
		t.Errorf(
			"Expected query.redirect_uri: \"%s\". Got: \"%s\"",
			expectedQuery.Get("redirect_uri"),
			gotQuery.Get("redirect_uri"),
		)
	}
}

func Test_GetOrganizations(t *testing.T) {
	tests := []struct {
		Title                 string
		ExistingOrganizations OpenFaaSCloudClaims
		ExpectedOrganizations []string
	}{
		{
			Title: "Example with properly separated organizations with comma",
			ExistingOrganizations: OpenFaaSCloudClaims{Organizations: "openfaas,openfaas-incubator,openfaas-cloud"},
			ExpectedOrganizations: []string{"openfaas", "openfaas-incubator", "openfaas-cloud"},
		}, {
			Title: "Example with un-proper separation of organizations with space",
			ExistingOrganizations: OpenFaaSCloudClaims{Organizations: "openfaas openfaas-incubator openfaas-cloud"},
			ExpectedOrganizations: []string{"openfaas openfaas-incubator openfaas-cloud"},
		},
	}
	for _, test := range tests {
		t.Run(test.Title, func(t *testing.T) {
			organizations := test.ExistingOrganizations.GetOrganizations()
			for order, value := range organizations {
				if test.ExpectedOrganizations[order] != value {
					t.Errorf("Expected: %s got: %s", test.ExpectedOrganizations[order], value)
				}
			}
		})
	}
}

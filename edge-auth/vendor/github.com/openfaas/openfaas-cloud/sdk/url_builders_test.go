package sdk

import (
	"testing"
)

func Test_GetSubdomain(t *testing.T) {
	tests := []struct {
		title             string
		URL               string
		expectedErr       error
		expectedSubdomain string
	}{
		{
			title:             "URL is with the right format",
			URL:               "https://www.organization.org",
			expectedErr:       nil,
			expectedSubdomain: "www",
		},
		{
			title:             "URL has only host",
			URL:               "www.organization.org",
			expectedErr:       nil,
			expectedSubdomain: "",
		},
		{
			title:             "URL is not set",
			URL:               "",
			expectedErr:       nil,
			expectedSubdomain: "",
		},
	}
	for _, test := range tests {
		Subdomain, domainErr := GetSubdomain(test.URL)
		if domainErr != test.expectedErr {
			t.Errorf("expected error: %v got: %v", test.expectedErr, domainErr)
		}
		if Subdomain != test.expectedSubdomain {
			t.Errorf("Expected URL: %s got: %s", test.expectedSubdomain, Subdomain)
		}
	}
}

func Test_FormatSystemURL(t *testing.T) {
	tests := []struct {
		title       string
		gatewayURL  string
		expectedURL string
	}{
		{
			title:       "\"gateway_public_url\" environmental variable is set with trailing slash",
			gatewayURL:  "https://cloud.o6s.io/",
			expectedURL: "https://system.o6s.io",
		},
		{
			title:       "\"gateway_public_url\" environmental variable is set without trailing slash",
			gatewayURL:  "https://cloud.o6s.io",
			expectedURL: "https://system.o6s.io",
		},
		{
			title:       "\"gateway_public_url\" environmental variable is unset",
			gatewayURL:  "",
			expectedURL: "system",
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			systemURL, _ := FormatSystemURL(test.gatewayURL)
			if systemURL != test.expectedURL {
				t.Errorf("Expected URL: %s got: %s", test.expectedURL, systemURL)
			}
		})
	}
}

func Test_FormatDashboardURL(t *testing.T) {
	tests := []struct {
		title       string
		event       *Event
		gatewayURL  string
		expectedURL string
	}{
		{
			title: "\"gateway_public_url\" environmental variable is set and Event object is set",
			event: &Event{
				Owner: "user",
			},
			gatewayURL:  "https://cloud.o6s.io/",
			expectedURL: "https://system.o6s.io/dashboard/user",
		},
		{
			title: "\"gateway_public_url\" environmental variable is unset and Event object is set",
			event: &Event{
				Owner: "user",
			},
			gatewayURL:  "",
			expectedURL: "system/dashboard/user",
		},
		{
			title:       "\"gateway_public_url\" environmental variable is unset and Event object is empty",
			event:       &Event{},
			gatewayURL:  "",
			expectedURL: "system/dashboard/",
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			dashboardURL, _ := FormatDashboardURL(test.gatewayURL, test.event)
			if dashboardURL != test.expectedURL {
				t.Errorf("Expected URL: %s got: %s", test.expectedURL, dashboardURL)
			}
		})
	}
}

func Test_FormatEndpointURL(t *testing.T) {
	tests := []struct {
		title       string
		event       *Event
		gatewayURL  string
		expectedURL string
	}{
		{
			title: "\"gateway_public_url\" environmental variable is set and Event object is set",
			event: &Event{
				Service: "cloud-func",
				Owner:   "user",
			},
			gatewayURL:  "https://cloud.o6s.io/",
			expectedURL: "https://user.o6s.io/cloud-func",
		},
		{
			title:       "\"gateway_public_url\" environmental variable is set but Event object is empty",
			event:       &Event{},
			gatewayURL:  "https://cloud.o6s.io",
			expectedURL: "https://.o6s.io/",
		},
		{
			title: "\"gateway_public_url\" environmental variable is unset but Event object is set",
			event: &Event{
				Service: "cloud-func",
				Owner:   "user",
			},
			gatewayURL:  "",
			expectedURL: "user/cloud-func",
		},
		{
			title:       "\"gateway_public_url\" environmental variable is empty and Event object is empty",
			event:       &Event{},
			gatewayURL:  "",
			expectedURL: "/",
		},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			endpointURL, _ := FormatEndpointURL(test.gatewayURL, test.event)
			if endpointURL != test.expectedURL {
				t.Errorf("Expected URL: %s got: %s", test.expectedURL, endpointURL)
			}
		})
	}
}
func Test_FormatLogsURL(t *testing.T) {
	tests := []struct {
		title       string
		event       *Event
		gatewayURL  string
		expectedURL string
	}{
		{
			title: "\"gateway_public_url\" environmental variable is configured and Event object is populated",
			event: &Event{
				Owner:      "user",
				Service:    "cloud-func",
				Repository: "cloud-func-repo",
				SHA:        "98zxc823axcc",
			},
			gatewayURL:  "https://cloud.o6s.io/",
			expectedURL: "https://system.o6s.io/dashboard/user/cloud-func/log?repoPath=user/cloud-func-repo&commitSHA=98zxc823axcc",
		},
		{
			title:       "\"gateway_public_url\" environmental variable is configured but the Event object is not populated",
			event:       &Event{},
			gatewayURL:  "https://cloud.o6s.io/",
			expectedURL: "https://system.o6s.io/dashboard///log?repoPath=/&commitSHA=",
		},
		{
			title: "\"gateway_public_url\" environmental variable is empty but the Event object is populated",
			event: &Event{
				Owner:      "user",
				Service:    "cloud-func",
				Repository: "cloud-func-repo",
				SHA:        "98zxc823axcc",
			},
			gatewayURL:  "",
			expectedURL: "system/dashboard/user/cloud-func/log?repoPath=user/cloud-func-repo&commitSHA=98zxc823axcc",
		},
		{
			title:       "\"gateway_public_url\" environmental variable is empty and Event object is not populated right",
			event:       &Event{},
			gatewayURL:  "",
			expectedURL: "system/dashboard///log?repoPath=/&commitSHA=",
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			logsURL, _ := FormatLogsURL(test.gatewayURL, test.event)
			if logsURL != test.expectedURL {
				t.Errorf("expected URL: %s got: %s", test.expectedURL, logsURL)
			}
		})
	}
}

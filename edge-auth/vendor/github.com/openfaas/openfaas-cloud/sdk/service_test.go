package sdk

import (
	"testing"
)

func Test_FormatServiceName(t *testing.T) {
	values := []struct {
		eventOwner   string
		functionName string
		expectedName string
	}{
		{
			eventOwner:   "ExampleName",
			functionName: "ExampleFunction",
			expectedName: "examplename-ExampleFunction",
		},
		{
			eventOwner:   "examplename",
			functionName: "ExampleFunction",
			expectedName: "examplename-ExampleFunction",
		},
		{
			eventOwner:   "examplename",
			functionName: "examplefunction",
			expectedName: "examplename-examplefunction",
		},
		{
			eventOwner:   "ExampleName",
			functionName: "examplefunction",
			expectedName: "examplename-examplefunction",
		},
	}
	for _, test := range values {
		serviceName := FormatServiceName(test.eventOwner, test.functionName)
		if serviceName != test.expectedName {
			t.Errorf("Expected name: `%v` got: `%v`", test.expectedName, serviceName)
		}
	}
}

func Test_CreateServiceURL(t *testing.T) {
	tests := []struct {
		title       string
		URL         string
		suffix      string
		expectedURL string
	}{
		{
			title:       "URL formatted for Swarm with port",
			URL:         "http://gateway:8080",
			suffix:      "",
			expectedURL: "http://gateway:8080",
		},
		{
			title:       "URL formatted for Kubernetes",
			URL:         "http://gateway:8080",
			suffix:      "openfaas",
			expectedURL: "http://gateway.openfaas:8080",
		},
		{
			title:       "URL formatted for Kubernetes showing backward compatability",
			URL:         "http://gateway.openfaas:8080",
			suffix:      "openfaas",
			expectedURL: "http://gateway.openfaas:8080",
		},
		{
			title:       "URL formatted for Kubernetes showing backward compatability when suffix is not present",
			URL:         "http://gateway.openfaas:8080",
			suffix:      "",
			expectedURL: "http://gateway.openfaas:8080",
		},
		{
			title:       "URL formatted for Swarm without port",
			URL:         "http://gateway",
			suffix:      "",
			expectedURL: "http://gateway",
		},
		{
			title:       "URL formatted for Kubernetes without port",
			URL:         "http://gateway",
			suffix:      "openfaas",
			expectedURL: "http://gateway.openfaas",
		},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			URL := CreateServiceURL(test.URL, test.suffix)
			if URL != test.expectedURL {
				t.Errorf("Expected: `%v` got: `%v`", test.expectedURL, URL)
			}
		})
	}
}

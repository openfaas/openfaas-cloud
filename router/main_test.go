package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func Test_makeHandler(t *testing.T) {
	config := []struct {
		RequestHost        string
		UpstreamURL        string
		ExpectedStatusCode int
	}{
		{
			RequestHost:        "martindekov.example.xyz/",
			UpstreamURL:        "http://localhost:8080/myfunction",
			ExpectedStatusCode: http.StatusInternalServerError,
		},
		{
			RequestHost:        "martindekov.example.xyz/",
			UpstreamURL:        "/myfunction",
			ExpectedStatusCode: http.StatusServiceUnavailable,
		},
		{
			RequestHost:        "",
			UpstreamURL:        "http://localhost:8080/myfunction",
			ExpectedStatusCode: http.StatusBadRequest,
		},
		{
			RequestHost:        "",
			UpstreamURL:        "http://localhost:8080/",
			ExpectedStatusCode: http.StatusBadRequest,
		},
	}

	timeout := time.Second * 30

	for _, test := range config {
		req, err := http.NewRequest(http.MethodPost, "/", nil)
		if err != nil {
			t.Fatalf("Request error: `%v`\n", err)
		}

		URI := strings.TrimPrefix(test.UpstreamURL, "http://localhost:8080")
		req.RequestURI = URI
		req.Host = test.RequestHost

		rec := httptest.NewRecorder()
		c := &http.Client{
			Timeout: timeout,
		}

		URL := strings.TrimSuffix(test.UpstreamURL, "myfunction")
		handlerFunc := makeHandler(c, timeout, URL)
		handlerFunc(rec, req)

		if rec.Code != test.ExpectedStatusCode {
			t.Errorf("Expected status code : `%v` got: `%v`", test.ExpectedStatusCode, rec.Code)
		}

		read, err := ioutil.ReadAll(rec.Body)
		if err != nil {
			t.Errorf("Unexpected response body read error: `%v`", err)
		}

		bodyValue := string(read)
		var ExpectedSuffix string
		userFromURI := strings.TrimPrefix(URI, "/")
		if len(test.RequestHost) == 0 {
			ExpectedSuffix = fmt.Sprintf("%s-%s", test.RequestHost, userFromURI)
		} else {
			functionName := test.RequestHost[0:strings.Index(test.RequestHost, ".")]
			ExpectedSuffix = fmt.Sprintf("%s-%s", functionName, userFromURI)
		}
		if !strings.Contains(bodyValue, ExpectedSuffix) {
			t.Errorf("The message: \n%v\nHad to contain: %v", bodyValue, ExpectedSuffix)
		}
	}
}

package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func Test_makeHandler(t *testing.T) {
	config := []struct {
		Title              string
		Host               string
		RequestURI         string
		ExpectedSuffix     string
		UpstreamURL        string
		Timeout            time.Duration
		ExpectedStatusCode int
	}{
		{
			Title:              "Right configuration with status 500 unrelated to function",
			Host:               "martindekov.example.xyz/",
			RequestURI:         "myfunction",
			ExpectedSuffix:     "martindekov-myfunction",
			UpstreamURL:        "http://localhost:8080/",
			Timeout:            time.Second * 30,
			ExpectedStatusCode: http.StatusInternalServerError,
		},
		{
			Title:              "Everything configured but upstream URL resulting in status 503",
			Host:               "martindekov.example.xyz/",
			RequestURI:         "myfunction",
			ExpectedSuffix:     "martindekov-myfunction",
			UpstreamURL:        "",
			Timeout:            time.Second * 30,
			ExpectedStatusCode: http.StatusServiceUnavailable,
		},
		{
			Title:              "Everything configured but Host resulting in status 400",
			Host:               "",
			RequestURI:         "/myfunction",
			ExpectedSuffix:     "-myfunction",
			UpstreamURL:        "http://localhost:8080",
			Timeout:            time.Second * 30,
			ExpectedStatusCode: http.StatusBadRequest,
		},
		{
			Title:              "Test with empty Host and URI resulting in bad request 400",
			Host:               "",
			RequestURI:         "",
			ExpectedSuffix:     "-",
			UpstreamURL:        "http://localhost:8080",
			Timeout:            time.Second * 30,
			ExpectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, test := range config {
		req, err := http.NewRequest(http.MethodPost, "/", nil)
		req.Host = test.Host
		req.RequestURI = test.RequestURI
		if err != nil {
			t.Logf("Test: `%v`\n", test.Title)
			t.Fatalf("\nRequest error: `%v`\n", err)
		}
		rec := httptest.NewRecorder()
		c := &http.Client{
			Timeout: test.Timeout,
		}
		handlerFunc := makeHandler(c, test.Timeout, test.UpstreamURL)
		handlerFunc(rec, req)
		res := rec.Result()
		defer res.Body.Close()
		read, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Logf("Test: `%v`\n", test.Title)
			t.Errorf("Unexpected responce body read error: `%v`", err)
		}
		stringBody := string(read)
		if res.StatusCode != test.ExpectedStatusCode {
			t.Logf("Test: `%v`\n", test.Title)
			t.Errorf("\nExpected status code : `%v` got: `%v`", test.ExpectedStatusCode, res.StatusCode)
		}
		if !strings.Contains(stringBody, test.ExpectedSuffix) {
			t.Logf("Test: `%v`\n", test.Title)
			t.Errorf("\nThe messege: \n%s\nHad to contain: %s", string(read), test.ExpectedSuffix)
		}
	}
}

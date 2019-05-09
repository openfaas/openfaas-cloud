package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

type gateway struct {
	RequestURI string
}

func (h *gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.RequestURI = r.URL.String()
	w.Write([]byte("\n"))
}

type passHandler struct {
	Next http.HandlerFunc
}

func (h passHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Next(w, r)
}

func Test_makeHandler(t *testing.T) {
	gatewayHandler := &gateway{}
	gateway := httptest.NewServer(gatewayHandler)
	defer gateway.Close()

	c := http.Client{}

	tests := []struct {
		Scenario           string
		RequestURL         string
		UpstreamURL        string
		ExpectedStatusCode int
	}{
		{
			Scenario:           "convert username to prefix",
			RequestURL:         "http://system.example.xyz/dashboard",
			UpstreamURL:        "/function/system-dashboard",
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Scenario:           "no sub-domain is invalid",
			RequestURL:         "http://example.xyz/dashboard",
			UpstreamURL:        "",
			ExpectedStatusCode: http.StatusBadRequest,
		},
		{
			Scenario:           "multiple function slash prefixes",
			RequestURL:         "http://system.example.xyz/////dashboard",
			UpstreamURL:        "/function/system-dashboard",
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Scenario:           "missing function name",
			RequestURL:         "http://system.example.xyz/",
			UpstreamURL:        "",
			ExpectedStatusCode: http.StatusNotFound,
		},
	}

	router := httptest.NewServer(passHandler{
		Next: makeHandler(&c, time.Second*10, gateway.URL, nil),
	})

	defer router.Close()

	for _, testCase := range tests {
		t.Run(testCase.Scenario, func(t *testing.T) {
			gatewayHandler.RequestURI = ""

			u, _ := url.Parse(testCase.RequestURL)

			req, _ := http.NewRequest(http.MethodGet, router.URL+u.Path, nil)
			req.Host = u.Host

			res, err := c.Do(req)
			if err != nil {
				t.Error(err)
				t.Fail()
			}

			if res.StatusCode != testCase.ExpectedStatusCode {
				t.Errorf("Status code want: %d, got %d", testCase.ExpectedStatusCode, res.StatusCode)
				t.Fail()
			}

			if gatewayHandler.RequestURI != testCase.UpstreamURL {

				t.Errorf("RequestURI want: %s, got: %s", testCase.UpstreamURL, gatewayHandler.RequestURI)
				t.Fail()
			}
		})
	}
}

package main

import (
	"fmt"
	"log"
	"net/http"
)

type authProxy struct {
	URL    string
	Client *http.Client
}

func (a *authProxy) Validate(upstreamURL string, r *http.Request) (int, string) {
	validateURL := a.URL + "q/?r=" + upstreamURL

	req, _ := http.NewRequest(http.MethodGet, validateURL, nil)

	for _, cookie := range r.Cookies() {
		req.AddCookie(cookie)
	}

	if len(r.Cookies()) == 0 {
		log.Println("No cookies to send.")
	} else {
		log.Printf("Cookies sent upstream: %d", len(req.Cookies()))
	}

	// Add all headers including referrer used for validation.
	copyHeaders(req.Header, &r.Header)

	res, err := a.Client.Do(req)

	if err != nil {
		log.Printf("Unable to reach auth service: %s", err.Error())
		return http.StatusBadGateway, ""
	}

	fmt.Println("Res:", res.Status)

	if res.Body != nil {
		defer res.Body.Close()
	}

	var location string
	locationURL, _ := res.Location()
	if locationURL != nil {
		location = locationURL.String()
	}

	log.Printf("Validating (%s) status: %d, location: %s\n", validateURL, res.StatusCode, location)

	return res.StatusCode, location
}

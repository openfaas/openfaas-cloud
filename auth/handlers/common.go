package handlers

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

func buildGitHubURL(config *Config, resource string, scope string) *url.URL {
	authURL := "https://github.com/login/oauth/authorize"
	u, _ := url.Parse(authURL)
	q := u.Query()

	q.Set("scope", scope)
	q.Set("allow_signup", "0")
	q.Set("state", fmt.Sprintf("%d", time.Now().Unix()))
	q.Set("client_id", config.ClientID)

	redirectURI := combineURL(config.ExternalRedirectDomain, "/oauth2/authorized")
	// if len(resource) > 0 {
	// 	redirectURI = redirectURI + "?r=" + resource
	// }

	q.Set("redirect_uri", redirectURI)

	u.RawQuery = q.Encode()
	return u
}

func combineURL(a, b string) string {
	if !strings.HasSuffix(a, "/") {
		a = a + "/"
	}
	if strings.HasPrefix(b, "/") {
		b = strings.TrimPrefix(b, "/")
	}

	return a + b
}

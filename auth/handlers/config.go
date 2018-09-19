package handlers

import (
	"time"
)

type Config struct {
	ClientID               string
	ClientSecret           string
	ExternalRedirectDomain string
	Scope                  string
	CookieRootDomain       string
	CookieExpiresIn        time.Duration // time.Hour * 48
}

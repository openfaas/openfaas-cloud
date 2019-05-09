package handlers

import (
	"time"
)

type Config struct {
	OAuthProvider          string
	OAuthProviderBaseURL   string
	ClientID               string
	ClientSecret           string
	OAuthClientSecretPath  string // OAuthClientSecretPath when given overrides the ClientSecret env-var
	ExternalRedirectDomain string
	Scope                  string
	CookieRootDomain       string
	CookieExpiresIn        time.Duration
	PublicKeyPath          string
	PrivateKeyPath         string
	Debug                  bool // Debug enables verbose logging of claims / cookies
}

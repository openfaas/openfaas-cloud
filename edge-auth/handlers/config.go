package handlers

import (
	"time"
)

type Config struct {
	OAuthProvider          string
	OAuthProviderBaseURL   string
	ClientID               string
	OAuthClientSecretPath  string
	ExternalRedirectDomain string
	Scope                  string
	CookieRootDomain       string
	CookieExpiresIn        time.Duration
	SecureCookie           bool
	PublicKeyPath          string
	PrivateKeyPath         string
	Debug                  bool // Debug enables verbose logging of claims / cookies
}

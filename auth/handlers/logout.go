package handlers

import (
	"net/http"
	"time"
)

// MakeLoginHandler creates a handler for logging out
func MakeLogoutHandler(config *Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			HttpOnly: true,
			Name:     cookieName,
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(0, 0),
			Domain:   config.CookieRootDomain,
		})

		w.WriteHeader(http.StatusOK)
	}
}

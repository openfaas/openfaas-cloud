package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

// MakeQueryHandler returns whether a client can access a resource
func MakeQueryHandler(config *Config, protected []string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		resource := query.Get("r")

		status := http.StatusOK
		if len(resource) == 0 {
			status = http.StatusBadRequest
		} else if isProtected(resource, protected) {
			if !validCookie(r, cookieName) {
				// status = http.StatusUnauthorized
				status = http.StatusTemporaryRedirect
				log.Printf("No cookie or an invalid cookie was found.\n")
			} else {
				log.Printf("A valid cookie was found.\n")
			}
		}

		log.Printf("Validate %s => %d\n", resource, status)

		if status == http.StatusTemporaryRedirect {
			redirect := buildGitHubURL(config, "", config.Scope)
			log.Printf("Go to %s\n", redirect.String())

			http.Redirect(w, r, redirect.String(), http.StatusTemporaryRedirect)
			return
		}
		w.WriteHeader(status)

	}
}

func isProtected(resource string, protected []string) bool {
	for _, prefix := range protected {
		if strings.HasPrefix(resource, prefix) {
			return true
		}
	}
	return false
}

func validCookie(r *http.Request, cookieName string) bool {
	fmt.Println(r.Cookies())
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return false
	}

	return len(cookie.Value) > 0
}

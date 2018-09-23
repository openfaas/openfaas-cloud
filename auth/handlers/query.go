package handlers

import (
	"crypto"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// MakeQueryHandler returns whether a client can access a resource
func MakeQueryHandler(config *Config, protected []string) func(http.ResponseWriter, *http.Request) {
	keydata, err := ioutil.ReadFile(config.PublicKeyPath)
	if err != nil {
		log.Fatal(err)
	}

	publicKey, keyErr := jwt.ParseECPublicKeyFromPEM(keydata)
	if keyErr != nil {
		log.Fatal("Load public key", keyErr)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		resource := query.Get("r")

		status := http.StatusOK
		if len(resource) == 0 {
			status = http.StatusBadRequest
		} else if isProtected(resource, protected) {
			started := time.Now()
			cookieStatus := validCookie(r, cookieName, publicKey)
			log.Printf("Cookie verified: %fs [%d]", time.Since(started).Seconds(), cookieStatus)

			if cookieStatus == http.StatusOK {
				status = http.StatusOK
			} else if cookieStatus == http.StatusNetworkAuthenticationRequired {
				// status = http.StatusUnauthorized
				status = http.StatusTemporaryRedirect
				log.Printf("No cookie or an invalid cookie was found.\n")
			} else {
				log.Printf("A valid cookie was found.\n")
				status = http.StatusUnauthorized
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

func validCookie(r *http.Request, cookieName string, publicKey crypto.PublicKey) int {
	fmt.Println("Cookies ", r.Cookies())

	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return http.StatusNetworkAuthenticationRequired
	}

	if len(cookie.Value) > 0 {
		log.Println("JWT ", cookie.Value)

		parsed, parseErr := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
			return publicKey, nil
		})

		if parseErr != nil {
			log.Println(parseErr)
			return http.StatusUnauthorized
		}

		if parsed.Valid {
			return http.StatusOK
		}
	}

	return http.StatusUnauthorized
}

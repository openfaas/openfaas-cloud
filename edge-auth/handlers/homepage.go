package handlers

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
)

type HomepageTokens struct {
	AccessToken string
	Login       string
}

// MakeHomepageHandler shows the homepage
func MakeHomepageHandler(config *Config) func(http.ResponseWriter, *http.Request) {
	keydata, err := ioutil.ReadFile(config.PublicKeyPath)
	if err != nil {
		log.Fatalf("unable to read path: %s, error: %s", config.PublicKeyPath, err.Error())
	}

	publicKey, keyErr := jwt.ParseECPublicKeyFromPEM(keydata)
	if keyErr != nil {
		log.Fatalf("unable to parse public key: %s", keyErr.Error())
	}

	return func(w http.ResponseWriter, r *http.Request) {

		cookie, err := r.Cookie(cookieName)
		if err != nil {
			log.Println("No cookie found.")
			http.Redirect(w, r, "/login/?r="+r.URL.Path, http.StatusTemporaryRedirect)
			return
		}

		parsed, parseErr := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
			return publicKey, nil
		})

		if parseErr != nil {
			log.Println(parseErr, cookie.Value)
			w.Write([]byte("Unable to decode cookie, please clear your cookies and sign-in again"))
			return
		}
		log.Printf("Parsed JWT: %v", parsed)

		tmpl, err := template.ParseFiles("./template/home.html")

		var tpl bytes.Buffer

		err = tmpl.Execute(&tpl, HomepageTokens{
			AccessToken: "Unavailable",
			Login:       "Unknown",
		})

		if err != nil {
			log.Panic("Error executing template: ", err)
		}

		w.Write(tpl.Bytes())
	}
}

package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type HomepageTokens struct {
	AccessToken string
	Login       string
}

// MakeHomepageHandler shows the homepage
func MakeHomepageHandler(config *Config) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Path)

		c, err := r.Cookie(cookieName)
		if err != nil {
			fmt.Println("No cookie found.")
			http.Redirect(w, r, "/login/?r="+r.URL.Path, http.StatusTemporaryRedirect)
			return
		}

		session := OpenFaaSCloudSession{}
		v, _ := base64.StdEncoding.DecodeString(c.Value)
		sessionErr := json.Unmarshal([]byte(v), &session)
		if sessionErr != nil {
			fmt.Println(sessionErr, c.Value)
			w.Write([]byte("Unable to decode cookie, please clear your cookies and sign-in again"))
			return
		}

		tmpl, err := template.ParseFiles("./template/home.html")

		var tpl bytes.Buffer

		log.Println("Token - ", v)

		err = tmpl.Execute(&tpl, HomepageTokens{
			AccessToken: session.GitHubAccessToken,
			Login:       session.Username,
		})

		if err != nil {
			log.Panic("Error executing template: ", err)
		}

		w.Write(tpl.Bytes())
	}
}

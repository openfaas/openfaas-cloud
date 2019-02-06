package handlers

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// MakeLoginHandler creates a handler for logging in
func MakeLoginHandler(config *Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Login to path", r.URL.Path)

		resource := "/"

		if val := r.URL.Query().Get("r"); len(val) > 0 {
			resource = val
		}

		if strings.EqualFold(r.URL.Path, "/login/github") {

			u := buildGitHubURL(config, resource, config.Scope)

			http.Redirect(w, r, u.String(), http.StatusTemporaryRedirect)
			return
		}

		contents, err := ioutil.ReadFile("./template/login.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(contents)
	}
}

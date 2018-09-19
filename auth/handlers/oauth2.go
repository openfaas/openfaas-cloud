package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/openfaas/openfaas-cloud/auth/provider"
)

// MakeOAuth2Handler makes a hanmdler for OAuth 2.0 redirects
func MakeOAuth2Handler(config *Config) func(http.ResponseWriter, *http.Request) {
	c := &http.Client{
		Timeout: time.Second * 3,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf(`OAuth 2 - "%s"`, r.URL.Path)
		if r.URL.Path != "/oauth2/authorized" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized OAuth callback."))
			return
		}

		if r.Body != nil {
			defer r.Body.Close()
		}

		reqQuery := r.URL.Query()
		code := reqQuery.Get("code")
		state := reqQuery.Get("state")
		if len(code) == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized OAuth callback, no code parameter given."))
			return
		}
		if len(state) == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized OAuth callback, no state parameter given."))
			return
		}

		fmt.Printf("Exchange: %s, for an access_token\n", code)
		tokenURL := "https://github.com/login/oauth/access_token"

		u, _ := url.Parse(tokenURL)
		q := u.Query()
		q.Set("client_id", config.ClientID)
		q.Set("client_secret", config.ClientSecret)

		q.Set("code", code)
		q.Set("state", state)

		u.RawQuery = q.Encode()
		fmt.Println("Posting to", u.String())

		req, _ := http.NewRequest(http.MethodPost, u.String(), nil)

		req.Header.Add("Accept", "application/json")
		res, err := c.Do(req)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error exchanging code for access_token"))

			log.Println(err)
			return
		}

		token, tokenErr := getToken(res)
		if tokenErr != nil {
			fmt.Println(tokenErr)
			return
		}

		session, err := createSession(c, token)

		sessionBytes, _ := json.Marshal(session)
		fmt.Println(string(sessionBytes))
		encodedCookie := base64.StdEncoding.EncodeToString(sessionBytes)

		http.SetCookie(w, &http.Cookie{
			HttpOnly: true,
			Name:     cookieName,
			Value:    encodedCookie,
			Path:     "/",
			Expires:  time.Now().Add(config.CookieExpiresIn),
			Domain:   config.CookieRootDomain,
		})

		fmt.Println("Redirect", reqQuery)

		// Redirect to original requested resource (if specified in r=)
		redirect := reqQuery.Get("r")
		if len(redirect) > 0 {
			log.Printf(`Found redirect value "r"=%s, instructing client to redirect.`, redirect)
			// http.Redirect(w, r, reqQuery.Get("r"), http.StatusTemporaryRedirect)

			w.Write([]byte(`<html><head></head>Redirecting.. <a href="redirect">to original resource</a>. <script>window.location.replace("` + redirect + `");</script></html>`))
			return
		}

		w.Write([]byte("You have been issued a cookie. Please navigate to the page you were looking for."))
	}
}

// OpenFaaSCloudSession is serialized in a cookie for the user
type OpenFaaSCloudSession struct {
	Sub               int    `json:"sub"`
	Username          string `json:"preferred_username"`
	Name              string `json:"name"`
	GitHubAccessToken string `json:"github_access_token"`
}

func createSession(c *http.Client, token GitHubAccessToken) (*OpenFaaSCloudSession, error) {
	var err error
	session := &OpenFaaSCloudSession{}

	client := provider.NewGitHub(c)
	profile, profileErr := client.GetProfile(token.AccessToken)
	if profileErr != nil {
		return session, profileErr
	}

	session.Sub = profile.ID
	session.Username = profile.Login
	session.Name = profile.Name
	session.GitHubAccessToken = token.AccessToken

	return session, err
}

func getToken(res *http.Response) (GitHubAccessToken, error) {
	token := GitHubAccessToken{}
	if res.Body != nil {
		defer res.Body.Close()

		tokenRes, _ := ioutil.ReadAll(res.Body)

		err := json.Unmarshal(tokenRes, &token)
		if err != nil {
			return token, err
		}
		return token, nil
	}

	return token, fmt.Errorf("no body received from server")
}

// GitHubAccessToken as issued by GitHub
type GitHubAccessToken struct {
	AccessToken string `json:"access_token"`
}

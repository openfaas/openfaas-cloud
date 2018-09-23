package handlers

import (
	"crypto"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/openfaas/openfaas-cloud/auth/provider"
)

// MakeOAuth2Handler makes a hanmdler for OAuth 2.0 redirects
func MakeOAuth2Handler(config *Config) func(http.ResponseWriter, *http.Request) {
	c := &http.Client{
		Timeout: time.Second * 3,
	}

	privateKeydata, err := ioutil.ReadFile(config.PrivateKeyPath)
	if err != nil {
		log.Fatal(err)
	}

	privateKey, keyErr := jwt.ParseECPrivateKeyFromPEM(privateKeydata)
	if keyErr != nil {
		log.Fatal("Load private key", keyErr)
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

		session, err := createSession(c, token, privateKey, config)
		if err != nil {
			log.Printf("Error creating session: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal server error creating JWT"))
			return
		}

		http.SetCookie(w, &http.Cookie{
			HttpOnly: true,
			Name:     cookieName,
			Value:    session,
			Path:     "/",
			Expires:  time.Now().Add(config.CookieExpiresIn),
			Domain:   config.CookieRootDomain,
		})

		log.Printf("SetCookie done, redirect to: %s", reqQuery)

		// Redirect to original requested resource (if specified in r=)
		redirect := reqQuery.Get("r")
		if len(redirect) > 0 {
			log.Printf(`Found redirect value "r"=%s, instructing client to redirect`, redirect)
			// http.Redirect(w, r, reqQuery.Get("r"), http.StatusTemporaryRedirect)

			w.Write([]byte(`<html><head></head>Redirecting.. <a href="redirect">to original resource</a>. <script>window.location.replace("` + redirect + `");</script></html>`))
			return
		}

		w.Write([]byte("You have been issued a cookie. Please navigate to the page you were looking for."))
	}
}

func createSession(c *http.Client, token GitHubAccessToken, privateKey crypto.PrivateKey, config *Config) (string, error) {
	var err error
	var session string

	client := provider.NewGitHub(c)
	profile, profileErr := client.GetProfile(token.AccessToken)
	if profileErr != nil {
		return session, profileErr
	}

	method := jwt.GetSigningMethod(jwt.SigningMethodES256.Name)
	claims := OpenFaaSCloudClaims{
		StandardClaims: jwt.StandardClaims{
			Id:        fmt.Sprintf("%d", profile.ID),
			Issuer:    "openfaas-cloud@github",
			ExpiresAt: time.Now().Add(48 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
			Subject:   profile.Login,
			Audience:  config.CookieRootDomain,
		},
		Name:        profile.Name,
		AccessToken: token.AccessToken,
	}

	session, err = jwt.NewWithClaims(method, claims).SignedString(privateKey)

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

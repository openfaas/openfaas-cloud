package handlers

import (
	"crypto"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/openfaas/openfaas-cloud/edge-auth/provider"
)

const profileFetchTimeout = time.Second * 5

// MakeOAuth2Handler makes a handler for OAuth 2.0 redirects
func MakeOAuth2Handler(config *Config) func(http.ResponseWriter, *http.Request) {
	c := &http.Client{
		Timeout: profileFetchTimeout,
	}

	privateKeydata, err := ioutil.ReadFile(config.PrivateKeyPath)
	if err != nil {
		log.Fatalf("private key, unable to read path: %s, error: %s", config.PrivateKeyPath, err.Error())
	}

	privateKey, keyErr := jwt.ParseECPrivateKeyFromPEM(privateKeydata)
	if keyErr != nil {
		log.Fatalf("unable to parse private key: %s", keyErr.Error())
	}

	clientSecret := config.ClientSecret

	if len(config.OAuthClientSecretPath) > 0 {
		clientSecretBytes, err := ioutil.ReadFile(config.OAuthClientSecretPath)
		if err != nil {
			log.Fatalf("OAuthClientSecretPath, unable to read path: %s, error: %s", config.OAuthClientSecretPath, err.Error())
		}
		clientSecret = strings.TrimSpace(string(clientSecretBytes))
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

		log.Printf("Exchange: %s, for an access_token", code)

		var tokenURL string
		var oauthProvider provider.Provider
		var redirectURI *url.URL

		switch config.OAuthProvider {
		case githubName:
			tokenURL = "https://github.com/login/oauth/access_token"
			oauthProvider = provider.NewGitHub(c)

			break
		case gitlabName:
			tokenURL = fmt.Sprintf("%s/oauth/token", config.OAuthProviderBaseURL)
			apiURL := config.OAuthProviderBaseURL + "/api/v4/"
			oauthProvider = provider.NewGitLabProvider(c, config.OAuthProviderBaseURL, apiURL)

			redirectAfterAutURL := reqQuery.Get("r")
			redirectURI, _ = url.Parse(combineURL(config.ExternalRedirectDomain, "/oauth2/authorized"))

			redirectURIQuery := redirectURI.Query()
			redirectURIQuery.Set("r", redirectAfterAutURL)

			redirectURI.RawQuery = redirectURIQuery.Encode()

			break
		}

		u, _ := url.Parse(tokenURL)
		q := u.Query()
		q.Set("client_id", config.ClientID)
		q.Set("client_secret", clientSecret)

		q.Set("code", code)
		q.Set("state", state)

		if config.OAuthProvider == gitlabName {
			q.Set("grant_type", "authorization_code")
			q.Set("redirect_uri", redirectURI.String())
		}

		u.RawQuery = q.Encode()
		log.Println("Posting to", u.String())

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
			log.Printf(
				"Unable to contact identity provider: %s, error: %s",
				config.OAuthProvider,
				tokenErr,
			)

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(
				"Unable to contact identity provider: %s",
				config.OAuthProvider,
			)))

			return
		}

		session, err := createSession(token, privateKey, config, oauthProvider, config.OAuthProvider)
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

			// Note: unable to redirect after setting Cookie, so landing on a redirect page instead.
			// http.Redirect(w, r, reqQuery.Get("r"), http.StatusTemporaryRedirect)

			w.Write([]byte(`<html><head></head>Redirecting.. <a href="redirect">to original resource</a>. <script>window.location.replace("` + redirect + `");</script></html>`))
			return
		}

		w.Write([]byte("You have been issued a cookie. Please navigate to the page you were looking for."))
	}
}

func createSession(token ProviderAccessToken, privateKey crypto.PrivateKey, config *Config, oauthProvider provider.Provider, providerName string) (string, error) {
	var err error
	var session string

	profile, profileErr := oauthProvider.GetProfile(token.AccessToken)
	if profileErr != nil {
		return session, profileErr
	}

	organizationList := ""

	if providerName == "github" {
		organizations, organizationsErr := getUserOrganizations(profile.Login, token.AccessToken)
		if organizationsErr != nil {
			return session, organizationsErr
		}
		organizationList = organizations
	}

	method := jwt.GetSigningMethod(jwt.SigningMethodES256.Name)
	claims := OpenFaaSCloudClaims{
		StandardClaims: jwt.StandardClaims{
			Id:        fmt.Sprintf("%d", profile.ID),
			Issuer:    fmt.Sprintf("openfaas-cloud@%s", config.OAuthProvider),
			ExpiresAt: time.Now().Add(48 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
			Subject:   profile.Login,
			Audience:  config.CookieRootDomain,
		},
		Organizations: organizationList,
		Name:          profile.Name,
		AccessToken:   token.AccessToken,
	}

	session, err = jwt.NewWithClaims(method, claims).SignedString(privateKey)

	return session, err
}

func getToken(res *http.Response) (ProviderAccessToken, error) {
	token := ProviderAccessToken{}
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

func getUserOrganizations(username, accessToken string) (string, error) {

	organizations := []Organization{}
	apiURL := fmt.Sprintf("https://api.github.com/users/%s/orgs", username)

	req, reqErr := http.NewRequest(http.MethodGet, apiURL, nil)
	if reqErr != nil {
		return "", fmt.Errorf("error while making request to `%s` organizations: %s", apiURL, reqErr.Error())
	}
	req.Header.Add("Authorization", "token "+accessToken)

	client := http.DefaultClient
	resp, respErr := client.Do(req)
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	if respErr != nil {
		return "", fmt.Errorf("error while requesting organizations: %s", respErr.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status code from request to GitHub organizations: %d", resp.StatusCode)
	}

	body, bodyErr := ioutil.ReadAll(resp.Body)
	if bodyErr != nil {
		return "", fmt.Errorf("error while reading body from GitHub organizations: %s", bodyErr.Error())
	}

	var allOrganizations []string
	unmarshallErr := json.Unmarshal(body, &organizations)
	if unmarshallErr != nil {
		return "", fmt.Errorf("error while un-marshaling organizations: %s", unmarshallErr.Error())
	}

	for _, organization := range organizations {
		allOrganizations = append(allOrganizations, organization.Login)
	}
	formatOrganizations := strings.Join(allOrganizations, ",")

	return formatOrganizations, nil
}

type Organization struct {
	Login string `json:"login"`
}

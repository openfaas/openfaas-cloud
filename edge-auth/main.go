package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/openfaas/openfaas-cloud/edge-auth/handlers"
	"github.com/openfaas/openfaas-cloud/edge-auth/provider"
)

const cookieExpiry = time.Hour * 48

func main() {
	var oauthProvider = "github"
	var oauthProviderBaseURL string

	var clientID string
	var clientSecret string
	var externalRedirectDomain string
	var cookieRootDomain string

	var publicKeyPath string
	var privateKeyPath string

	var oauthClientSecretPath string

	var writeDebug bool

	if val, exists := os.LookupEnv("oauth_provider"); exists {
		oauthProvider = val
	}

	if !provider.IsSupported(oauthProvider) {
		log.Fatalf(
			"OAuth 2 provider %s is not supported. Currently supported providers: %s",
			oauthProvider,
			provider.GetSupportedString(),
		)
	}

	if val, exists := os.LookupEnv("oauth_provider_base_url"); exists {
		oauthProviderBaseURL = val
	}

	if val, exists := os.LookupEnv("client_id"); exists {
		clientID = val
	}

	if val, exists := os.LookupEnv("client_secret"); exists {
		clientSecret = val
	}

	if val, exists := os.LookupEnv("external_redirect_domain"); exists {
		externalRedirectDomain = val
	}

	if val, exists := os.LookupEnv("cookie_root_domain"); exists {
		cookieRootDomain = val
	}

	if val, exists := os.LookupEnv("public_key_path"); exists {
		publicKeyPath = val
	}

	if val, exists := os.LookupEnv("private_key_path"); exists {
		privateKeyPath = val
	}

	if val, exists := os.LookupEnv("oauth_client_secret_path"); exists {
		oauthClientSecretPath = val
	}

	if val, exists := os.LookupEnv("write_debug"); exists && (val == "true" || val == "1") {
		writeDebug = true
	}

	config := &handlers.Config{
		OAuthProvider:          strings.ToLower(oauthProvider),
		OAuthProviderBaseURL:   oauthProviderBaseURL,
		ClientID:               clientID,
		ClientSecret:           clientSecret,
		CookieExpiresIn:        cookieExpiry,
		CookieRootDomain:       cookieRootDomain,
		ExternalRedirectDomain: externalRedirectDomain,
		Scope:                  "read:org,read:user,user:email",
		PublicKeyPath:          publicKeyPath,
		PrivateKeyPath:         privateKeyPath,
		OAuthClientSecretPath:  oauthClientSecretPath,
		Debug:                  writeDebug,
	}

	protected := []string{
		"/function/system-dashboard",
		"/function/system-list-functions",
		"/function/system-metrics",
	}

	fs := http.FileServer(http.Dir("static"))

	router := http.NewServeMux()
	router.Handle("/static/", http.StripPrefix("/static/", fs))

	router.HandleFunc("/", handlers.MakeHomepageHandler(config))

	router.HandleFunc("/q/", handlers.MakeQueryHandler(config, protected))
	router.HandleFunc("/login/", handlers.MakeLoginHandler(config))
	router.HandleFunc("/oauth2/", handlers.MakeOAuth2Handler(config))
	router.HandleFunc("/healthz/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK."))
	})

	timeout := time.Second * 10
	port := 8080
	if v, exists := os.LookupEnv("port"); exists {
		val, _ := strconv.Atoi(v)
		port = val
	}

	log.Printf("Using port: %d\n", port)

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		Handler:        router,
		ReadTimeout:    timeout,
		WriteTimeout:   timeout,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(s.ListenAndServe())
}

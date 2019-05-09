package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const authHost = "auth.system"

func main() {
	cfg := NewRouterConfig()

	if len(cfg.UpstreamURL) == 0 {
		log.Panicln("give an upstream_url as an env-var")
	}

	if len(cfg.AuthURL) == 0 {
		log.Panicln("give an auth_url as an env-var")
	}

	maxIdleConns := 1024
	maxIdleConnsPerHost := 1024

	proxyClient := makeProxy(cfg.Timeout, maxIdleConns, maxIdleConnsPerHost)

	log.Printf("Timeout set to: %s\n", cfg.Timeout)
	log.Printf("Upstream URL: %s\n", cfg.UpstreamURL)

	authProxy1 := authProxy{
		URL:    cfg.AuthURL,
		Client: proxyClient,
	}

	router := http.NewServeMux()
	router.HandleFunc("/", makeHandler(proxyClient, cfg.Timeout, cfg.UpstreamURL, &authProxy1))
	router.HandleFunc("/healthz", makeHealthzHandler())

	log.Printf("Using port %s\n", cfg.Port)

	s := &http.Server{
		Addr:           ":" + cfg.Port,
		Handler:        router,
		ReadTimeout:    cfg.Timeout,
		WriteTimeout:   cfg.Timeout,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(s.ListenAndServe())
}

type authProxy struct {
	URL    string
	Client *http.Client
}

func (a *authProxy) Validate(upstreamURL string, r *http.Request) (int, string) {
	validateURL := a.URL + "q/?r=" + upstreamURL

	req, _ := http.NewRequest(http.MethodGet, validateURL, nil)

	for _, cookie := range r.Cookies() {
		req.AddCookie(cookie)
	}

	if len(r.Cookies()) == 0 {
		log.Println("No cookies to send.")
	} else {
		log.Printf("Cookies sent upstream: %d", len(req.Cookies()))
	}

	// Add all headers including referrer used for validation.
	copyHeaders(req.Header, &r.Header)

	res, err := a.Client.Do(req)

	if err != nil {
		log.Printf("Unable to reach auth service: %s", err.Error())
		return http.StatusBadGateway, ""
	}

	fmt.Println("Res:", res.Status)

	if res.Body != nil {
		defer res.Body.Close()
	}

	var location string
	locationURL, _ := res.Location()
	if locationURL != nil {
		location = locationURL.String()
	}

	log.Printf("Validating (%s) status: %d, location: %s\n", validateURL, res.StatusCode, location)

	return res.StatusCode, location
}

// makeHandler builds a router to convert sub-domains into OpenFaaS gateway URLs with
// a username prefix and suffix of the destination function.
// i.e. system.o6s.io/dashboard
//      becomes: gateway:8080/function/system-dashboard, where gateway:8080
//      is specified in upstreamURL
func makeHandler(c *http.Client, timeout time.Duration, upstreamURL string, auth *authProxy) func(w http.ResponseWriter, r *http.Request) {

	if strings.HasSuffix(upstreamURL, "/") == false {
		upstreamURL = upstreamURL + "/"
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		var host string

		tldSepCount := 1
		tldSep := "."
		if len(r.Host) == 0 || strings.Count(r.Host, tldSep) <= tldSepCount {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid sub-domain in Host header"))
			return
		}

		host = r.Host[0:strings.Index(r.Host, tldSep)]
		fmt.Printf("Router host: %s (%s)\n", host, r.Host)

		requestURI := r.RequestURI
		requestURI = strings.TrimLeft(requestURI, "/")

		if len(requestURI) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var upstreamFullURL *url.URL

		isAuthHost := strings.HasPrefix(r.Host, authHost)
		if isAuthHost {
			var err error
			upstreamFullURL, err = url.Parse(fmt.Sprintf("%s%s", auth.URL, requestURI))
			if err != nil {
				log.Printf("Auth URL transparent error: %s\n", err)
			} else {
				log.Printf("Auth URL transparent %s\n", upstreamFullURL.String())
			}
		} else {
			upstreamFullURL, _ = url.Parse(fmt.Sprintf("%sfunction/%s-%s", upstreamURL, host, requestURI))
		}

		if auth != nil && !isAuthHost {
			authStatus, location := auth.Validate(upstreamFullURL.Path, r)
			fmt.Println(authStatus, location)

			responseWritten := false
			switch authStatus {
			case http.StatusUnauthorized:
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Unauthorized"))

				responseWritten = true
				break
			case http.StatusTemporaryRedirect:

				directTo, _ := url.Parse(location)
				q := directTo.Query()

				returnTo := "http://" + r.Host + "" + r.RequestURI

				redirectURI, _ := url.Parse(q.Get("redirect_uri"))
				log.Printf(`Redirect URL: "%s"\n`, redirectURI)

				redirectURIQuery := redirectURI.Query()
				redirectURIQuery.Set("r", returnTo)

				redirectURI.RawQuery = redirectURIQuery.Encode()

				log.Printf(`* Redirect URL: "%s"\n`, redirectURI)
				q.Set("redirect_uri", redirectURI.String())

				directTo.RawQuery = q.Encode()

				log.Println("Go to: ", r.RequestURI, r.URL.String())

				log.Printf("Auth caused redirect to: %s\n", directTo.String())
				http.Redirect(w, r, directTo.String(), http.StatusTemporaryRedirect)
				responseWritten = true
				break
			case http.StatusBadGateway:
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("bad gateway reaching auth server"))
				responseWritten = true
				break
			case http.StatusOK:
				log.Printf("Auth cleared. OK.\n")
				break
			}

			if responseWritten {
				return
			}
		}

		req, _ := http.NewRequest(r.Method, upstreamFullURL.String(), r.Body)

		timeoutContext, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		copyHeaders(req.Header, &r.Header)

		log.Printf("Serving: %s\n", req.URL.String())

		res, resErr := c.Do(req.WithContext(timeoutContext))
		if resErr != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(resErr.Error()))

			fmt.Printf("Upstream %s status: %d\n", upstreamFullURL, http.StatusBadGateway)
			return
		}

		copyHeaders(w.Header(), &res.Header)
		fmt.Printf("Upstream %s status: %d\n", upstreamFullURL, res.StatusCode)

		w.WriteHeader(res.StatusCode)
		if res.Body != nil {
			defer res.Body.Close()

			bytesOut, _ := ioutil.ReadAll(res.Body)
			w.Write(bytesOut)
		}
	}
}

func copyHeaders(destination http.Header, source *http.Header) {
	for k, v := range *source {
		vClone := make([]string, len(v))
		copy(vClone, v)
		(destination)[k] = vClone
	}
}

func makeProxy(timeout time.Duration, maxIdleConns, maxIdleConnsPerHost int) *http.Client {
	client := http.DefaultClient
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	client.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: timeout,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          maxIdleConns,
		MaxIdleConnsPerHost:   maxIdleConnsPerHost,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return client
}

func makeHealthzHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
			break

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

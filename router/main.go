package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

func main() {
	cfg := NewRouterConfig()

	if len(cfg.UpstreamURL) == 0 {
		log.Panicln("give an upstream_url as an env-var")
	}

	c := makeProxy(cfg.Timeout)

	log.Printf("Timeout set to: %s\n", cfg.Timeout)
	log.Printf("Upstream URL: %s\n", cfg.UpstreamURL)

	router := http.NewServeMux()

	router.HandleFunc("/", makeHandler(c, cfg.Timeout, cfg.UpstreamURL))

	s := &http.Server{
		Addr:           ":" + cfg.Port,
		Handler:        router,
		ReadTimeout:    cfg.Timeout,
		WriteTimeout:   cfg.Timeout,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(s.ListenAndServe())
}

func makeHandler(c *http.Client, timeout time.Duration, upstreamURL string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		if len(r.Host) == 0 {
			w.WriteHeader(http.StatusBadRequest)
		}

		requestURI := r.RequestURI
		if strings.HasPrefix(requestURI, "/") {
			requestURI = requestURI[1:]
		}

		path := fmt.Sprintf("%sfunction/%s-%s", upstreamURL, r.Host[:strings.Index(r.Host, ".")], requestURI)

		fmt.Printf("Proxying to: %s\n", path)

		if r.Body != nil {
			defer r.Body.Close()
		}
		req, _ := http.NewRequest(r.Method, path, r.Body)

		timeoutContext, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		copyHeaders(req.Header, &r.Header)

		res, resErr := c.Do(req.WithContext(timeoutContext))
		if resErr != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(resErr.Error()))

			fmt.Printf("Upstream %s status: %d\n", path, http.StatusBadGateway)
			return
		}

		copyHeaders(w.Header(), &res.Header)
		fmt.Printf("Upstream %s status: %d\n", path, res.StatusCode)

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

func makeProxy(timeout time.Duration) *http.Client {
	proxyClient := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   timeout,
				KeepAlive: 1 * time.Second,
			}).DialContext,
			IdleConnTimeout:       120 * time.Millisecond,
			ExpectContinueTimeout: 1500 * time.Millisecond,
		},
	}
	return &proxyClient
}

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	port := "8080"

	if portVal, exists := os.LookupEnv("port"); exists && len(portVal) > 0 {
		port = portVal
	}

	var upstreamURL string
	if up, exists := os.LookupEnv("upstream_url"); exists && len(up) > 0 {
		if strings.HasSuffix(up, "/") == false {
			up = up + "/"
		}

		upstreamURL = up
	}

	if len(upstreamURL) == 0 {
		log.Panicln("give an upstream_url as an env-var")
	}

	timeout := time.Second * 60
	c := makeProxy(timeout)
	router := http.NewServeMux()
	router.HandleFunc("/", makeHandler(c, upstreamURL))

	s := &http.Server{
		Addr:           ":" + port,
		Handler:        router,
		ReadTimeout:    timeout,
		WriteTimeout:   timeout,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(s.ListenAndServe())
}

func makeHandler(c *http.Client, upstreamURL string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		if len(r.Host) == 0 {
			w.WriteHeader(http.StatusBadRequest)
		}

		requestURI := r.RequestURI
		if strings.HasPrefix(requestURI, "/") {
			requestURI = requestURI[1:]
		}

		path := fmt.Sprintf("%sfunction/%s-%s", upstreamURL, r.Host[0:strings.Index(r.Host, ".")], requestURI)

		fmt.Printf("Proxying to: %s\n", path)

		if r.Body != nil {
			defer r.Body.Close()
		}
		req, _ := http.NewRequest(r.Method, path, r.Body)

		copyHeaders(req.Header, &r.Header)

		res, resErr := c.Do(req)
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
			// MaxIdleConns:          1,
			// DisableKeepAlives:     false,
			IdleConnTimeout:       120 * time.Millisecond,
			ExpectContinueTimeout: 1500 * time.Millisecond,
		},
	}
	return &proxyClient
}

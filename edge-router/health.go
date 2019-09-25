package main

import (
	"net/http"
)

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

package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func Test_makeHandler(t *testing.T) {
	body := ""
	req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("Error is: `%v`", err)
	}

	rec := httptest.NewRecorder()
	client := &http.Client{}
	timeout := time.Second * 30

	handlerFnc := makeHandler(client, timeout, "localost:8080/wat")

	handlerFnc(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("wat")
	}
	read, _ := ioutil.ReadAll(rec.Body)
	val := string(read)
	t.Log(val)

}

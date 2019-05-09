package main

import (
	"os"
	"testing"
)

func TestReadConfig_PortOverride(t *testing.T) {
	want := "8081"
	os.Setenv("port", want)

	cfg := NewRouterConfig()
	got := cfg.Port

	if want != got {
		t.Errorf("want port %s, but got: %s", want, got)
		t.Fail()
	}
}

func TestReadConfig_Defaults(t *testing.T) {
	want := "8080"
	os.Setenv("port", "")

	cfg := NewRouterConfig()
	got := cfg.Port

	if want != got {
		t.Errorf("want port %s, but got: %s", want, got)
		t.Fail()
	}
}

func TestReadConfig_UpstreamURLGiven(t *testing.T) {
	want := "http://127.0.0.1:31111/"
	os.Setenv("upstream_url", want)

	cfg := NewRouterConfig()
	got := cfg.UpstreamURL

	if want != got {
		t.Errorf("want upstream_url %s, but got: %s", want, got)
		t.Fail()
	}
}

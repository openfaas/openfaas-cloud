package handlers

import "testing"

func TestCombineURL_BuildsValidURL(t *testing.T) {
	want := "https://www.google.com/query/this"
	got := combineURL("https://www.google.com/query/", "/this")

	if want != got {
		t.Errorf("combineURL want: %s, got: %s", want, got)
		t.Fail()
	}
}

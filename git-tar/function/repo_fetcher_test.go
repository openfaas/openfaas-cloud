package function

import (
	"os"
	"strings"
	"testing"

	"github.com/openfaas/openfaas-cloud/sdk"
)

func Test_Clone(t *testing.T) {
	fetcher := FakeFetcher{}
	ev := sdk.PushEvent{
		Repository: sdk.PushEventRepository{
			Owner: sdk.Owner{
				Login: "alexellis",
			},
			Name: "test-repo",
		},
	}

	path, err := clone(fetcher, ev)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	wantPathSuffix := ev.Repository.Owner.Login + "-" + ev.Repository.Name
	if strings.HasSuffix(path, wantPathSuffix) {
		t.Errorf("want suffix: %s, path was: %s", wantPathSuffix, path)
	}

	if _, err := os.Stat(path); err != nil {
		t.Errorf("repo not cloned to: %s", path)
	}

	defer os.RemoveAll(path)
}

func Test_CloneErrorsWithEmptyOwner(t *testing.T) {
	fetcher := FakeFetcher{}
	ev := sdk.PushEvent{
		Repository: sdk.PushEventRepository{
			Owner: sdk.Owner{
				Login: "",
			},
			Name: "test-repo",
		},
	}

	_, err := clone(fetcher, ev)
	if err == nil {
		t.Error("no login should throw an error")
		t.Fail()
		return
	}

	wantErr := "login must be specified"
	if err.Error() != wantErr {
		t.Errorf("want err: %q, got %q", wantErr, err.Error())
		t.Fail()
	}
}

func Test_CloneErrorsWithEmptyRepoName(t *testing.T) {
	fetcher := FakeFetcher{}
	ev := sdk.PushEvent{
		Repository: sdk.PushEventRepository{
			Owner: sdk.Owner{
				Login: "alexellis",
			},
			Name: "",
		},
	}

	_, err := clone(fetcher, ev)
	if err == nil {
		t.Error("no repo name should throw an error")
		t.Fail()
		return
	}

	wantErr := "repo name must be specified"
	if err.Error() != wantErr {
		t.Errorf("want err: %q, got %q", wantErr, err.Error())
		t.Fail()
	}
}

type FakeFetcher struct {
}

func (c FakeFetcher) Clone(url, path string) error {
	return os.MkdirAll(path, 0700)
}

func (c FakeFetcher) Checkout(commitID, path string) error {
	return os.MkdirAll(path, 0700)
}

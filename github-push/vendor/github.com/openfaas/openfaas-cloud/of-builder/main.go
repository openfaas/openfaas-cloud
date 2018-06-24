package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/docker/docker/pkg/archive"
	"github.com/gorilla/mux"
	"github.com/moby/buildkit/util/appcontext"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

func main() {
	flag.Parse()

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/build", buildHandler)

	server := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: router,
	}

	eg, ctx := errgroup.WithContext(appcontext.Context())

	eg.Go(func() error {
		<-ctx.Done()
		return server.Shutdown(context.Background())
	})

	eg.Go(func() error {
		return server.ListenAndServe()
	})

	if err := eg.Wait(); err != nil {
		panic(err)
	}
}

func buildHandler(w http.ResponseWriter, r *http.Request) {
	dt, err := build(w, r)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("%s", err.Error())))
		return
	}
	w.WriteHeader(200)
	w.Write(dt)
}

func build(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	tmpdir, err := ioutil.TempDir("", "buildctx")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpdir)

	if err := archive.Untar(r.Body, tmpdir, nil); err != nil {
		return nil, err
	}

	dt, err := ioutil.ReadFile(filepath.Join(tmpdir, "config"))
	if err != nil {
		return nil, err
	}

	var cfg struct {
		Ref      string
		Frontend string
	}

	if err := json.Unmarshal(dt, &cfg); err != nil {
		return nil, err
	}

	if cfg.Ref == "" {
		return nil, errors.Errorf("no target reference to push")
	}

	if cfg.Frontend == "" {
		cfg.Frontend = "tonistiigi/dockerfile:v0"
	}

	cmd := exec.Command("buildctl", "--debug", "build", "--frontend=gateway.v0", "--frontend-opt=source="+cfg.Frontend, "--local=context="+filepath.Join(tmpdir, "context"), "--local=dockerfile="+filepath.Join(tmpdir, "context"), "--no-progress", "--exporter=image",
		"--exporter-opt=name="+cfg.Ref, "--exporter-opt=push=true", "--exporter-opt=registry.insecure=true")
	env := os.Environ()
	env = append(env, "BUILDKIT_HOST=tcp://of-buildkit:1234")
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return []byte(cfg.Ref), nil
}

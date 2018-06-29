package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/docker/docker/pkg/archive"
	"github.com/gorilla/mux"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth/authprovider"
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
	contextDir := filepath.Join(tmpdir, "context")
	solveOpt := client.SolveOpt{
		Exporter: "image",
		ExporterAttrs: map[string]string{
			"name":              cfg.Ref,
			"push":              "true",
			"registry.insecure": "true",
		},
		LocalDirs: map[string]string{
			"context":    contextDir,
			"dockerfile": contextDir,
		},
		Frontend: "dockerfile.v0",
		// FrontendAttrs: map[string]string{
		// 	"source": cfg.Frontend,
		// },
		// ~/.docker/config.json could be provided as Kube or Swarm's secret
		Session: []session.Attachable{authprovider.NewDockerAuthProvider()},
	}
	c, err := client.New("tcp://of-buildkit:1234", client.WithBlock())
	if err != nil {
		return nil, err
	}
	ch := make(chan *client.SolveStatus)
	eg, ctx := errgroup.WithContext(context.Background())
	eg.Go(func() error {
		return c.Solve(ctx, nil, solveOpt, ch)
	})
	eg.Go(func() error {
		for s := range ch {
			for _, v := range s.Vertexes {
				log.Printf("vertex: %s %s %v %v", v.Digest, v.Name, v.Started, v.Completed)
			}
			for _, s := range s.Statuses {
				log.Printf("status: %s %s %d", s.Vertex, s.ID, s.Current)
			}
			for _, l := range s.Logs {
				log.Printf("log: %s\n%s", l.Vertex, l.Data)
			}
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return []byte(cfg.Ref), nil
}

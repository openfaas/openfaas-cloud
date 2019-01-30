package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/alexellis/hmac"
	"github.com/docker/docker/pkg/archive"
	"github.com/gorilla/mux"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth/authprovider"
	"github.com/moby/buildkit/util/appcontext"
	"github.com/openfaas/openfaas-cloud/sdk"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

var (
	lchownEnabled bool
	buildkitURL   string
	buildArgs     = map[string]string{}
)

func main() {
	flag.Parse()

	lchownEnabled = true
	if val, exists := os.LookupEnv("enable_lchown"); exists {
		if val == "false" {
			lchownEnabled = false
		}
	}

	buildkitURL = "tcp://of-buildkit:1234"
	if val, ok := os.LookupEnv("buildkit_url"); ok && len(val) > 0 {
		buildkitURL = val
	}

	if val, ok := os.LookupEnv("http_proxy"); ok {
		buildArgs["build-arg:http_proxy"] = val
	}

	if val, ok := os.LookupEnv("https_proxy"); ok {
		buildArgs["build-arg:https_proxy"] = val
	}

	if val, ok := os.LookupEnv("no_proxy"); ok {
		buildArgs["build-arg:no_proxy"] = val
	}

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/build", buildHandler)
	router.HandleFunc("/healthz", healthzHandler)

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

	dt, err := build(w, r, buildArgs)

	if err != nil {
		w.WriteHeader(500)

		if dt == nil {
			buildResult := BuildResult{
				ImageName: "",
				Log:       nil,
				Status:    fmt.Sprintf("unexpected failure: %s", err.Error()),
			}
			dt, _ = json.Marshal(buildResult)
		}
		w.Write(dt)

		// w.Write([]byte(fmt.Sprintf("%s", err.Error())))
		return
	}
	w.WriteHeader(200)
	w.Write(dt)
}

func build(w http.ResponseWriter, r *http.Request, buildArgs map[string]string) ([]byte, error) {

	if r.Body == nil {
		return nil, fmt.Errorf("a body is required to build a function")
	}

	defer r.Body.Close()

	tmpdir, err := ioutil.TempDir("", "buildctx")
	if err != nil {
		return nil, err
	}

	tarBytes, bodyErr := ioutil.ReadAll(r.Body)
	if bodyErr != nil {
		return nil, bodyErr
	}

	hmacErr := validateRequest(&tarBytes, r)

	if hmacErr != nil {
		return nil, hmacErr
	}

	defer os.RemoveAll(tmpdir)

	opts := archive.TarOptions{
		NoLchown: !lchownEnabled,
	}

	if err := archive.Untar(bytes.NewReader(tarBytes), tmpdir, &opts); err != nil {
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

	insecure := "false"
	if val, exists := os.LookupEnv("insecure"); exists {
		insecure = val
	}

	frontendAttrs := map[string]string{
		"source": cfg.Frontend,
	}

	for k, v := range buildArgs {
		frontendAttrs[k] = v
	}

	contextDir := filepath.Join(tmpdir, "context")
	solveOpt := client.SolveOpt{
		Exporter: "image",
		ExporterAttrs: map[string]string{
			"name": strings.ToLower(cfg.Ref),
			"push": "true",
		},
		LocalDirs: map[string]string{
			"context":    contextDir,
			"dockerfile": contextDir,
		},
		Frontend:      "dockerfile.v0",
		FrontendAttrs: frontendAttrs,
		// ~/.docker/config.json could be provided as Kube or Swarm's secret
		Session: []session.Attachable{authprovider.NewDockerAuthProvider()},
	}

	if insecure == "true" {
		solveOpt.ExporterAttrs["registry.insecure"] = insecure
	}

	c, err := client.New(buildkitURL, client.WithBlock())
	if err != nil {
		return nil, err
	}

	ch := make(chan *client.SolveStatus)
	eg, ctx := errgroup.WithContext(context.Background())
	eg.Go(func() error {
		return c.Solve(ctx, nil, solveOpt, ch)
	})

	build := buildLog{
		Line: []string{},
		Sync: &sync.Mutex{},
	}

	eg.Go(func() error {
		for s := range ch {
			for _, v := range s.Vertexes {
				var msg string
				if v.Completed != nil {
					msg = fmt.Sprintf("v: %s %s %.2fs", v.Started.Format(time.RFC3339), v.Name, v.Completed.Sub(*v.Started).Seconds())
				} else {
					var startedTime time.Time
					if v.Started != nil {
						startedTime = *(v.Started)
					} else {
						startedTime = time.Now()
					}
					startedVal := startedTime.Format(time.RFC3339)
					msg = fmt.Sprintf("v: %s %v", startedVal, v.Name)
				}
				build.Append(msg)
				fmt.Printf("%s\n", msg)

			}
			for _, s := range s.Statuses {
				msg := fmt.Sprintf("s: %s %s %d", s.Timestamp.Format(time.RFC3339), s.ID, s.Current)
				build.Append(msg)

				fmt.Printf("status: %s %s %d\n", s.Vertex, s.ID, s.Current)
			}
			for _, l := range s.Logs {

				msg := fmt.Sprintf("l: %s %s", l.Timestamp.Format(time.RFC3339), l.Data)
				build.Append(msg)

				fmt.Printf("log: %s\n%s\n", l.Vertex, l.Data)
			}

		}
		return nil
	})

	if err := eg.Wait(); err != nil {

		buildResult := BuildResult{
			ImageName: cfg.Ref,
			Log:       build.Line,
			Status:    fmt.Sprintf("failure: %s", err.Error()),
		}

		bytesOut, _ := json.Marshal(buildResult)
		return bytesOut, err
	}

	buildResult := BuildResult{
		ImageName: cfg.Ref,
		Log:       build.Line,
		Status:    "success",
	}

	bytesOut, _ := json.Marshal(buildResult)

	return bytesOut, nil
}

// BuildResult represents a successful Docker build and
// push operation to a remote registry
type BuildResult struct {
	Log       []string `json:"log"`
	ImageName string   `json:"imageName"`
	Status    string   `json:"status"`
}

type buildLog struct {
	Line []string
	Sync *sync.Mutex
}

func (b *buildLog) Append(msg string) {
	b.Sync.Lock()
	defer b.Sync.Unlock()

	b.Line = append(b.Line, msg)

}

func validateRequest(req *[]byte, r *http.Request) (err error) {
	payloadSecret, err := sdk.ReadSecret("payload-secret")

	if err != nil {
		return fmt.Errorf("couldn't get payload-secret: %t", err)
	}

	xCloudSignature := r.Header.Get(sdk.CloudSignatureHeader)

	err = hmac.Validate(*req, xCloudSignature, payloadSecret)

	if err != nil {
		return err
	}

	return nil
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodGet:
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		break

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

module github.com/openfaas/openfaas-cloud/of-builder

go 1.13

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78
	github.com/Microsoft/go-winio v0.4.7
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5
	github.com/alexellis/hmac v0.0.0-20180624210714-d5d71edd7bc7
	github.com/containerd/continuity v0.0.0-20180416230128-c6cef3483023
	github.com/docker/cli v0.0.0-20180503173406-0ff5f5205159
	github.com/docker/distribution v2.6.0-rc.1.0.20170825220652-30578ca32960+incompatible
	github.com/docker/docker v1.4.2-0.20180506231517-5f395b35bc60
	github.com/docker/docker-credential-helpers v0.6.0
	github.com/docker/go-connections v0.3.0
	github.com/docker/go-units v0.3.3
	github.com/gogo/protobuf v1.0.0
	github.com/golang/protobuf v1.1.0
	github.com/google/shlex v0.0.0-20150127133951-6f45313302b9
	github.com/gorilla/context v0.0.0-20160226214623-1ea25387ff6f
	github.com/gorilla/mux v1.6.1
	github.com/moby/buildkit v0.0.0-20180507051859-fabec2957873
	github.com/opencontainers/go-digest v1.0.0-rc1
	github.com/opencontainers/image-spec v1.0.1
	github.com/opencontainers/runc v0.1.1
	github.com/openfaas/faas-provider v0.0.0-20180910095832-845bf7aa58cb
	github.com/openfaas/openfaas-cloud v0.0.0-20180927141003-6abeccfcf77b
	github.com/opentracing/opentracing-go v1.0.2
	github.com/pkg/errors v0.8.0
	github.com/sirupsen/logrus v1.0.2-0.20170713114250-a3f95b5c4235
	github.com/tonistiigi/fsutil v0.0.0-20180414035453-93a0fd10b669
	github.com/tonistiigi/grpc-opentracing v0.0.0-20180106052059-420e5c3331a0
	golang.org/x/net v0.0.0-20180502164142-640f4622ab69
	golang.org/x/sync v0.0.0-20180314180146-1d60e4601c6f
	golang.org/x/sys v0.0.0-20180504064212-6f686a352de6
	golang.org/x/text v0.3.0
	google.golang.org/genproto v0.0.0-20180427144745-86e600f69ee4
	google.golang.org/grpc v1.11.3
)

replace github.com/grpc-ecosystem/grpc-opentracing 420e5c3331a082e0c873adeeab1819fe1749dd0b => github.com/tonistiigi/grpc-opentracing v0.0.0-20180106052059-420e5c3331a0

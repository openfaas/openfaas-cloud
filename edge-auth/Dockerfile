FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.13 as build

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

WORKDIR /go/src/github.com/openfaas/openfaas-cloud/edge-auth

ENV CGO_ENABLED=0
ENV GO111MODULE=off

COPY vendor     vendor
COPY handlers   handlers
COPY static     static
COPY template   template
COPY provider   provider
COPY main.go    .

# Run a gofmt and exclude all vendored code.
RUN test -z "$(gofmt -l $(find . -type f -name '*.go' -not -path "./vendor/*"))" || { echo "Run \"gofmt -s -w\" on your Golang code"; exit 1; }

RUN CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} go test -v

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=${CGO_ENABLED} go build \
        --ldflags "-s -w" \
        -a -installsuffix cgo \
        -o edge-auth .

FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine:3.12 as ship

RUN apk --no-cache add ca-certificates \
    && addgroup -S app && adduser -S -g app app \
    && mkdir -p /home/app \
    && chown app /home/app

WORKDIR /home/app/

COPY --from=build /go/src/github.com/openfaas/openfaas-cloud/edge-auth/edge-auth /bin/
COPY --from=build /go/src/github.com/openfaas/openfaas-cloud/edge-auth/static      static
COPY --from=build /go/src/github.com/openfaas/openfaas-cloud/edge-auth/template    template

LABEL org.opencontainers.image.source https://github.com/openfaas/openfaas-cloud

USER app
EXPOSE 8080
VOLUME /tmp

CMD ["edge-auth"]

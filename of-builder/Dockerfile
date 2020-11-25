FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.13 as build

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

WORKDIR /go/src/github.com/openfaas/openfaas-cloud/of-builder

ENV CGO_ENABLED=0
ENV GO111MODULE=off

#RUN apk add --no-cache git g++ linux-headers curl ca-certificates

RUN curl -sSLf https://amazon-ecr-credential-helper-releases.s3.us-east-2.amazonaws.com/0.3.1/linux-amd64/docker-credential-ecr-login > docker-credential-ecr-login \
 && chmod +x docker-credential-ecr-login \
 && mv docker-credential-ecr-login /usr/local/bin/

# Run a gofmt and exclude all vendored code.
RUN test -z "$(gofmt -l $(find . -type f -name '*.go' -not -path "./vendor/*"))" || { echo "Run \"gofmt -s -w\" on your Golang code"; exit 1; }

ADD main.go     .
ADD healthz.go  .
ADD vendor      vendor

RUN CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} go test -v

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=${CGO_ENABLED} go build \
        --ldflags "-s -w" \
        -a -installsuffix cgo \
        -o of-builder .

FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine:3.12 as ship

RUN apk --no-cache add ca-certificates \
    && addgroup -S app && adduser -S -g app app \
    && mkdir -p /home/app \
    && chown app /home/app

RUN mkdir -p /home/app/.aws/

WORKDIR /home/app

COPY --from=build /usr/local/bin/docker-credential-ecr-login  /usr/local/bin/
COPY --from=build /go/src/github.com/openfaas/openfaas-cloud/of-builder/of-builder .

RUN chown -R app /home/app

USER app
EXPOSE 8080
VOLUME /tmp/

ENTRYPOINT ["./of-builder"]

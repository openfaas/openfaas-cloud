FROM golang:1.11-alpine AS builder

WORKDIR /go/src/github.com/openfaas/openfaas-cloud/edge-router

ENV CGO_ENABLED=0

COPY main.go            .
COPY main_test.go       .
COPY config.go          .
COPY config_test.go     .
COPY health.go          .
COPY auth_proxy.go      .

RUN go test -v \
    && go build -o /usr/bin/edge-router .

FROM alpine:3.10

RUN apk add --no-cache ca-certificates

COPY --from=builder /usr/bin/edge-router /bin/

EXPOSE 8080

VOLUME /tmp

ENTRYPOINT ["edge-router"]

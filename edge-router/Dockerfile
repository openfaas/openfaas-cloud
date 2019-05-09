FROM golang:1.10-alpine AS builder

WORKDIR /go/src/github.com/openfaas/openfaas-cloud/edge-router

COPY main.go            .
COPY main_test.go       .
COPY config.go          .
COPY config_test.go     .

RUN go test -v \
    && go build -o /usr/bin/edge-router .

FROM alpine:3.9

RUN apk add --no-cache ca-certificates

COPY --from=builder /usr/bin/edge-router /bin/

EXPOSE 8080

VOLUME /tmp

ENTRYPOINT ["edge-router"]

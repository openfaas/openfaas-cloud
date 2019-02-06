FROM golang:1.10-alpine AS builder
RUN apk add --no-cache g++
WORKDIR /go/src/github.com/openfaas/openfaas-cloud/router

COPY main.go            .
COPY main_test.go       .
COPY config.go          .
COPY config_test.go     .

RUN go test -v \
    && go build -o /usr/bin/router .

FROM alpine:3.9

RUN apk add --no-cache ca-certificates

COPY --from=builder /usr/bin/router /bin/

EXPOSE 8080

VOLUME /tmp

ENTRYPOINT ["router"]

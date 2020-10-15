FROM golang:1.13-alpine AS build

WORKDIR /go/src/github.com/openfaas/openfaas-cloud/edge-router

ENV CGO_ENABLED=0
ENV GO111MODULE=off

COPY main.go            .
COPY main_test.go       .
COPY config.go          .
COPY config_test.go     .
COPY health.go          .
COPY auth_proxy.go      .

RUN go test -v \
    && go build -o /usr/bin/edge-router .

FROM alpine:3.12 as ship

RUN apk --no-cache add ca-certificates \
    && addgroup -S app && adduser -S -g app app \
    && mkdir -p /home/app \
    && chown app /home/app

COPY --from=build /usr/bin/edge-router /bin/

WORKDIR /home/app/
USER app
EXPOSE 8080
VOLUME /tmp

ENTRYPOINT ["edge-router"]

FROM golang:1.13-alpine AS build

WORKDIR /go/src/github.com/openfaas/openfaas-cloud/edge-auth

ENV CGO_ENABLED=0
ENV GO111MODULE=off

COPY vendor     vendor
COPY handlers   handlers
COPY static     static
COPY template   template
COPY provider   provider
COPY main.go    .

RUN go test -v \
    && go build -o /usr/bin/edge-auth .

FROM alpine:3.12 as ship
RUN apk --no-cache add ca-certificates \
    && addgroup -S app && adduser -S -g app app \
    && mkdir -p /home/app \
    && chown app /home/app

WORKDIR /home/app/
COPY --from=build /usr/bin/edge-auth /bin/
COPY --from=build /go/src/github.com/openfaas/openfaas-cloud/edge-auth/static      static
COPY --from=build /go/src/github.com/openfaas/openfaas-cloud/edge-auth/template    template

USER app
EXPOSE 8080
VOLUME /tmp

CMD ["edge-auth"]

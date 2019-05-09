FROM golang:1.10-alpine AS builder

WORKDIR /go/src/github.com/openfaas/openfaas-cloud/edge-auth

COPY vendor     vendor
COPY handlers   handlers
COPY static     static
COPY template   template
COPY provider   provider
COPY main.go    .

RUN go test -v \
    && go build -o /usr/bin/edge-auth .

FROM alpine:3.9
RUN apk add --no-cache ca-certificates

WORKDIR /root/
COPY --from=builder /usr/bin/edge-auth /bin/
COPY --from=builder /go/src/github.com/openfaas/openfaas-cloud/edge-auth/static      static
COPY --from=builder /go/src/github.com/openfaas/openfaas-cloud/edge-auth/template    template

EXPOSE 8080

VOLUME /tmp

CMD ["edge-auth"]

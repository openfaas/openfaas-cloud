FROM golang:1.11-alpine AS builder

WORKDIR /go/src/github.com/openfaas/openfaas-cloud/edge-auth

ENV CGO_ENABLED=0

COPY vendor     vendor
COPY handlers   handlers
COPY static     static
COPY template   template
COPY provider   provider
COPY main.go    .

RUN go test -v \
    && go build -o /usr/bin/edge-auth .

FROM alpine:3.10
RUN apk add --no-cache ca-certificates

WORKDIR /root/
COPY --from=builder /usr/bin/edge-auth /bin/
COPY --from=builder /go/src/github.com/openfaas/openfaas-cloud/edge-auth/static      static
COPY --from=builder /go/src/github.com/openfaas/openfaas-cloud/edge-auth/template    template

EXPOSE 8080

VOLUME /tmp

CMD ["edge-auth"]

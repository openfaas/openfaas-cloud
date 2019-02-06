FROM golang:1.10-alpine AS builder
RUN apk add --no-cache g++
WORKDIR /go/src/github.com/openfaas/openfaas-cloud/auth

COPY vendor     vendor
COPY handlers   handlers
COPY static     static
COPY template   template
COPY provider   provider
COPY main.go    .

RUN go test -v \
    && go build -o /usr/bin/auth .

FROM alpine:3.9
RUN apk add --no-cache ca-certificates

WORKDIR /root/
COPY --from=builder /usr/bin/auth /bin/
COPY --from=builder /go/src/github.com/openfaas/openfaas-cloud/auth/static      static
COPY --from=builder /go/src/github.com/openfaas/openfaas-cloud/auth/template    template

EXPOSE 8080

VOLUME /tmp

CMD ["auth"]

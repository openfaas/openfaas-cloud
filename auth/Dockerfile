FROM golang:1.9-alpine AS builder

WORKDIR /go/src/github.com/openfaas/openfaas-cloud/auth

COPY handlers   handlers
COPY static     static
COPY template   template
COPY provider   provider
COPY main.go    .

RUN go test -v \
    && go build -o /usr/bin/auth .

FROM alpine:3.8
RUN apk add --no-cache ca-certificates

WORKDIR /root/
COPY --from=builder /usr/bin/auth /bin/
COPY --from=builder /go/src/github.com/openfaas/openfaas-cloud/auth/static      static
COPY --from=builder /go/src/github.com/openfaas/openfaas-cloud/auth/template    template

EXPOSE 8080

VOLUME /tmp

CMD ["auth"]

FROM golang:1.9-alpine AS builder

WORKDIR /go/src/github.com/openfaas/openfaas-cloud/router

COPY main.go .
# COPY vendor  vendor

RUN go build -o /usr/bin/router .

FROM alpine:3.7
COPY --from=builder /usr/bin/router /bin/

EXPOSE 8080

VOLUME /tmp

ENTRYPOINT ["router"]
FROM golang:1.11-alpine AS builder

ENV CGO_ENABLED=0

RUN apk add --no-cache git g++ linux-headers

WORKDIR /go/src/github.com/openfaas/openfaas-cloud/of-builder

ADD main.go .
ADD healthz.go .
ADD vendor  vendor

RUN go build -o /usr/bin/of-builder .

FROM alpine:3.10

# Setting the group prevented access to /tmp at runtime
# lchown started to fail
# -G app 
RUN addgroup -S app && adduser app -S \
 && mkdir -p /home/app

WORKDIR /home/app

COPY --from=builder /usr/bin/of-builder /home/app/

RUN chown -R app /home/app
USER app

EXPOSE 8080
VOLUME /tmp/

ENTRYPOINT ["/home/app/of-builder"]

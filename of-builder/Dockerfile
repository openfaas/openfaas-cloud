FROM golang:1.9-alpine AS builder
RUN apk add --no-cache git g++ linux-headers
WORKDIR /go/src/github.com/alexellis/of-cloud/of-builder
ADD main.go .
ADD vendor  vendor

RUN go build -o /usr/bin/of-builder .

FROM tonistiigi/buildkit:standalone
COPY --from=builder /usr/bin/of-builder /bin/

EXPOSE 8080
VOLUME /tmp

ENTRYPOINT ["of-builder"]
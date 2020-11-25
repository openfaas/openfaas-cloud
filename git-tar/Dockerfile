FROM --platform=${TARGETPLATFORM:-linux/amd64} ghcr.io/openfaas/faas-cli:0.12.19 as faas-cli
FROM --platform=${TARGETPLATFORM:-linux/amd64} ghcr.io/openfaas/classic-watchdog:0.1.4 as watchdog
FROM --platform=${TARGETPLATFORM:-linux/amd64} golang:1.13-alpine3.12 as build

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

ENV CGO_ENABLED=0
ENV GO111MODULE=off

COPY --from=watchdog /fwatchdog /usr/bin/
COPY --from=faas-cli /usr/bin/faas-cli /usr/bin/

WORKDIR /go/src/handler
COPY . .

# Run a gofmt and exclude all vendored code.
RUN test -z "$(gofmt -l $(find . -type f -name '*.go' -not -path "./vendor/*"))" \
    || { echo "Run \"gofmt -s -w\" on your Golang code"; exit 1; }

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build --ldflags "-s -w" -a -installsuffix cgo -o handler .
RUN go test $(go list ./... | grep -v /vendor/) -cover

FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine:3.12 as ship

RUN apk --no-cache add \
    ca-certificates \
    libarchive-tools \
    git

# Add non root user
RUN addgroup -S app && adduser -S -g app app \
   && mkdir -p /home/app \
   && chown app /home/app

ENV cgi_headers=true
ENV combine_output=true

WORKDIR /home/app

COPY --from=build /go/src/handler/handler    .
COPY --from=build /usr/bin/fwatchdog         /usr/bin/fwatchdog
COPY --from=build /usr/bin/faas-cli          /usr/local/bin/faas-cli

RUN chmod 777 /tmp

USER app

ENV fprocess="./handler"

HEALTHCHECK --interval=5s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]

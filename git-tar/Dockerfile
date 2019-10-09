FROM golang:1.11-alpine3.10 as build

ENV CGO_ENABLED=0

RUN apk --no-cache add curl \ 
    && echo "Pulling binaries from GitHub." \
    && curl -sSLf https://github.com/openfaas/faas/releases/download/0.18.1/fwatchdog > /usr/bin/fwatchdog \
    && curl -sSLf https://github.com/openfaas/faas-cli/releases/download/0.9.3/faas-cli > /usr/local/bin/faas-cli \
    && chmod +x /usr/bin/fwatchdog \
    && chmod +x /usr/local/bin/faas-cli \
    && apk del curl --no-cache

WORKDIR /go/src/handler
COPY . .

# Run a gofmt and exclude all vendored code.
RUN test -z "$(gofmt -l $(find . -type f -name '*.go' -not -path "./vendor/*"))" \
    || { echo "Run \"gofmt -s -w\" on your Golang code"; exit 1; }

RUN CGO_ENABLED=0 GOOS=linux \
    go build --ldflags "-s -w" -a -installsuffix cgo -o handler . && \
    go test $(go list ./... | grep -v /vendor/) -cover

FROM alpine:3.10

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
COPY --from=build /usr/local/bin/faas-cli    /usr/local/bin/faas-cli

RUN chmod 777 /tmp

USER app

ENV fprocess="./handler"

HEALTHCHECK --interval=5s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]

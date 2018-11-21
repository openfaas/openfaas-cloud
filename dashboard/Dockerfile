FROM node:10.12.0-alpine as build

RUN apk --no-cache add curl \
    && echo "Pulling watchdog binary from Github." \
    && curl -sSLf https://github.com/openfaas-incubator/of-watchdog/releases/download/0.4.0/of-watchdog > /usr/bin/fwatchdog \
    && chmod +x /usr/bin/fwatchdog

RUN mkdir -p /home/root/client

WORKDIR /home/root/client/

# Turn down the verbosity to default level.
ENV NPM_CONFIG_LOGLEVEL warn

COPY client/package.json ./

RUN npm i

COPY client/src ./src
COPY client/public ./public

RUN npm run build

### Final stage

FROM node:10.12.0-alpine

WORKDIR /root/

RUN addgroup -S -g 998 app && adduser -S -u 998 -g app app

RUN mkdir -p /home/app/function

WORKDIR /home/app/function/

# Turn down the verbosity to default level.
ENV NPM_CONFIG_LOGLEVEL warn

# This ordering means the npm installation is cached for the outer function handler.
COPY of-cloud-dashboard/package.json ./

RUN npm i

COPY of-cloud-dashboard/handler.js ./

# Copy built dashboard client from stage 1
COPY --from=build /home/root/client/build/ ./dist/

# Copy watchdog from stage 1
COPY --from=build /usr/bin/fwatchdog /usr/bin/fwatchdog

WORKDIR /home/app/

# This ordering means the npm installation is cached for the outer function handler.
COPY package.json ./

RUN npm i

# Copy outer function handler
COPY index.js ./

# chmod for tmp is for a buildkit issue (@alexellis)
RUN chown app:app -R /home/app \
    && chmod 777 /tmp

USER app

ENV cgi_headers="true"
ENV fprocess="node index.js"
ENV mode="http"
ENV upstream_url="http://127.0.0.1:3000"

ENV exec_timeout="20s"
ENV write_timeout="25s"
ENV read_timeout="25s"

HEALTHCHECK --interval=1s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]
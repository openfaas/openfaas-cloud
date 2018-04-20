#!/bin/sh
export SQUASH=true

cd ..

(cd of-builder && make) && \
(cd router && make) && \
(cd sdk && go build) && \

CLI="faas-cli"
if ! [ -x "$(command -v faas-cli)" ]; then
    HERE=`pwd`
    cd /tmp/
    curl -sL cli.openfaas.com | sh
    CLI="/tmp/faas-cli"
    cd $HERE
fi

$CLI build --parallel=4

#!/bin/sh

# Run this command from one-level higher in the folder path, not this folder.

(cd of-builder && make) && \
(cd router && make build) && \
(cd sdk && go build) && \

CLI="faas-cli"
if ! [ -x "$(command -v faas-cli)" ]; then
    HERE=`pwd`
    cd /tmp/
    curl -sL https://github.com/openfaas/faas-cli/releases/download/0.6.9/faas-cli > faas-cli
    chmod +x ./faas-cli
    CLI="/tmp/faas-cli"

    echo "Returning to $HERE"
    cd $HERE
fi

echo "Working folder: `pwd`"

$CLI build --parallel=4
# 
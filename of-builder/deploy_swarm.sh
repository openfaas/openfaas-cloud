#!/bin/sh

# Running with an unprotected local registry is not recommended

# docker service rm registry
# docker service create --network func_functions \
#   --name registry \
#   --detach=true -p 5000:5000 registry:latest

docker rm -f of-buildkit
docker run -d --net func_functions -d --privileged \
--restart always \
--name of-buildkit moby/buildkit:v0.3.3 --addr tcp://0.0.0.0:1234

export OF_BUILDER_TAG=0.6.2

# You should mount your .docker/config.json file here, but first make sure it is
# readable. `chmod 777 $HOME/.docker/config.json`

docker service create --constraint="node.role==manager" \
 --name of-builder \
 --env insecure=false --detach=true --network func_functions \
 --secret src=registry-secret,target="/home/app/.docker/config.json" \
 --secret src=payload-secret,target="/var/openfaas/secrets/payload-secret" \
 --env enable_lchown=false \
openfaas/of-builder:$OF_BUILDER_TAG

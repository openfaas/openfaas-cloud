#!/bin/sh

docker service rm registry
docker service create --network func_functions \
  --name registry \
  --detach=true -p 5000:5000 registry:latest

docker rm -f of-buildkit
docker run -d --net func_functions -d --privileged \
--restart always \
--name of-buildkit alexellis2/buildkit:2018-04-17 --addr tcp://0.0.0.0:1234

export OF_BUILDER_TAG=0.4.2

docker service rm of-builder
docker service create --constraint="node.role==manager" --detach=true \
 --network func_functions --name of-builder openfaas/of-builder:$OF_BUILDER_TAG



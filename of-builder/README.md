# Setup

When deploying OpenFaaS make sure you update the network to "attachable" in the docker-compose.yml file:

```
networks:
  functions:
    driver: overlay
    attachable: true
```

Now setup the registry and builder:

```
docker service rm registry
docker service create --network func_functions --name registry --detach=true -p 5000:5000  registry:latest

docker rm -f of-builder
docker build -t alexellis2/of-builder:0.2 .
docker run -d --net func_functions --name of-builder --privileged alexellis2/of-builder:0.2

rm req.tar

# prepare request tar
echo >config<<EOT                                                                                                        
{"Ref": "registry.local:5000/foo/bar:latest"}
EOT

mkdir -p context

cat >context/Dockerfile<<EOT                                                                                            
FROM busybox
ADD Dockerfile /
ENV foo bar
EOT

tar cvf req.tar .

# query
curl -i 192.168.10.98:8080/build -X POST --data-binary @req.tar
```

Outside of OpenFaaS:

```
docker build -t alexellis2/of-builder .

docker network create builder --driver overlay --attachable=true ; \
 docker service rm registry; \
 docker service create --network builder --name registry --detach=true -p 5000:5000  registry:latest ; \
 docker rm -f of-builder ; \
docker run -p 8080:8080 -d --net builder --name of-builder --privileged alexellis2/of-builder
```

Test:

```

docker rm -f dind; docker run --name dind --privileged --net=builder -d docker:dind dockerd --insecure-registry registry:5000
docker exec -ti dind docker pull registry:5000/jmkhael/figlet:latest-99745ca9f5a1a914384686e0e928a10854cc87d5
```


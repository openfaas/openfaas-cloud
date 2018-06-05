# of-builder

of-builder is an image builder for OpenFaaS images, it needs to be deployed with a container registry.

The following instructions are for Docker Swarm but this code will work with Kubernetes when configured correctly.

## Setup

Before you start deploy OpenFaaS via https://docs.openfaas.com/

* Setup the registry:

```
docker service rm registry
docker service create --network func_functions \
  --name registry \
  --detach=true -p 5000:5000 registry:latest
```

Warning: this exposes the registry without authentication on port 5000 publicly on your host. By binding to the local machine we can then pull images without using --insecure-registry specifying a 127.0.0.1 IP address on Linux, or your machine's IP on Docker for Mac.

* Run the buildkit daemon

buildkit will build Docker images from a tar-ball and push them to a registry

Note: note daemon contains a gRPC endpoint listening on port 1234.

```
docker rm -f of-buildkit
docker run -d --net func_functions -d --privileged \
--restart always \
--name of-buildkit akihirosuda/buildkit-rootless:20180605 --addr tcp://0.0.0.0:1234
```

* Setup the builder service:

The builder service calls into the buildkit daemon to build an OpenFaaS function over HTTP.

* Build

```
cd of-builder/
docker rm -f of-builder
export TAG=0.4.0
make

make push
```

* Deploy

```
export TAG=0.4.0
docker service create --network func_functions --name of-builder openfaas/of-builder:$TAG
```

## Do a test build (optional)

If you're testing the of-builder service you will need to publish the port 8080 as some non-conflicting value.

I.e.

```
docker service update of-builder --publish-add 8088:8080
```

We specify a config file which is JSON and tells buildkit which image to publish to. In this example it's going to be `registry.local:5000/foo/bar:latest`. The container image has a README.md added into it as an example. It also has an env-var set up.

```
rm -rf test-image && \
mkdir -p test-image && \
cd test-image

echo '{"Ref": "registry.local:5000/foo/bar:latest"}' > config

mkdir -p context
echo "## Made with buildkit" >> context/README.md
cat >context/Dockerfile<<EOT
FROM busybox
ADD README.md /
ENV foo bar
EOT

tar cvf req.tar  --exclude=req.tar  .
```

If `registry.local:5000` gives a DNS issue on Docker for Mac then change this to the public IP of your computer instead i.e. 192.168.0.100:5000.

## Post the tar to the builder

Change the IP as required

```
curl -i localhost:8088/build -X POST --data-binary @req.tar
```

## Test the image

To test the image just type in `docker run -ti 127.0.0.1:5000/foo/bar:latest cat /README.md` for instance.

## Appendix

### Testing without OpenFaaS:

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


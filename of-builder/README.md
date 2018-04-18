# Setup

Now setup the registry and builders:

```
docker service rm registry
docker service create --network func_functions --name registry --detach=true -p 5000:5000  registry:latest
docker run -d --net func_functions -d --privileged --name of-buildkit alexellis2/buildkit:2018-04-17 --addr tcp://0.0.0.0:1234

cd of-builder/
docker rm -f of-builder
docker build -t alexellis2/of-builder:0.3 .
docker run -d --net func_functions -p 8088:8080 --name of-builder alexellis2/of-builder:0.3
```

# Do a test build

We specify a config file which is JSON and tells buildkit which image to publish to. In this example it's going to be `registry.local:5000/foo/bar:latest`. The container image has a README.md added into it as an example. It also has an env-var set up.

```
mkdir image
cd image

echo '{"Ref": "registry.local:5000/foo/bar:latest"}' > config

mkdir -p context
echo "## Made with buildkit" >> context/README.md
cat >context/Dockerfile<<EOT
FROM busybox
ADD README.md /
ENV foo bar
EOT

tar cvf req.tar .
```

If `registry.local:5000` gives a DNS issue on Docker for Mac then change this to the public IP of your computer instead i.e. 192.168.0.100:5000.

# Post the tar to the builder

Change the IP as required

```
curl -i localhost:8088/build -X POST --data-binary @req.tar
```

# Test the image

To test the image just type in `docker run -ti 127.0.0.1:5000/foo/bar:latest cat /README.md` for instance.

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


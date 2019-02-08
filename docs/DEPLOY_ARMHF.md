# Deploying openfaas cloud on armhf

## Prerequisites

1. A domain name which can point to your deployment. (subdomain names also work)
2. Setup Github app. Refer [this](https://docs.openfaas.com/openfaas-cloud/self-hosted/github/). You'll need to transfer the downloaded `private-key.pem` file to arm device.
3. Docker hub account.

## Core components steps:
1. Sign in to docker hub on your machine using:

```bash
DOCKER_USERNAME=(enter your docker username)
DOCKER_PASSWORD=(enter your docker password)
docker login --username $DOCKER_USERNAME --password $DOCKER_PASSWORD
```

2. You'll need to set 'registry-secret' and 'payload-secret' using command
```bash
cat $HOME/.docker/config.json | docker secret create registry-secret -

PAYLOAD_SECRET=$(head -c 12 /dev/urandom | shasum| cut -d' ' -f1)
echo -n "$PAYLOAD_SECRET" | docker secret create payload-secret -
```

3. Run buildkit and of-builder with
```bash
docker rm -f of-buildkit
docker run -d --net func_functions -d --privileged \
--restart always \
--name of-buildkit moby/buildkit:v0.3.3 --addr tcp://0.0.0.0:1234

docker service create --constraint="node.role==manager" \
 --name of-builder \
 --env insecure=false --detach=true --network func_functions \
 --secret src=registry-secret,target="/home/app/.docker/config.json" \
 --secret src=payload-secret,target="/var/openfaas/secrets/payload-secret" \
 --env enable_lchown=false \
zeerorg/of-builder:armhf
```
4. Setup Github app secrets

```bash
WEBHOOK_SECRET="Long-Password-Phrase-Goes-Here"
echo -n "$WEBHOOK_SECRET" | docker secret create github-webhook-secret -

docker secret create private-key "location of private-key.pem"
```

5. Deploy minio using:
```bash
SECRET_KEY=$(head -c 12 /dev/urandom | shasum| cut -d' ' -f1)
ACCESS_KEY=$(head -c 12 /dev/urandom | shasum| cut -d' ' -f1)
docker service rm minio

docker service create --constraint="node.role==manager" \
 --name minio \
 --detach=true --network func_functions \
 --secret s3-access-key \
 --secret s3-secret-key \
 --env MINIO_SECRET_KEY_FILE=s3-secret-key \
 --env MINIO_ACCESS_KEY_FILE=s3-access-key \
zeerorg/minio-armhf:latest server /export
```

6. Change `gateway_config.yml` and remove all occurences of `.openfaas` (standard step for docker swarm)

7. `faas-cli deploy -f stack.armhf.yml` and you should be good to go


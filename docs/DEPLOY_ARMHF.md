# Deploying openfaas cloud on armhf

## Prerequisites

1. A domain name which can point to your deployment. (subdomain names also work)
2. Setup Github app. Refer [this](https://docs.openfaas.com/openfaas-cloud/self-hosted/github/). You'll need to transfer the downloaded `private-key.pem` file to arm device.
3. Docker hub account.

## Core components steps (with docker swarm):
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

7. Change `gateway_config.yml`
 - remove all occurences of `.openfaas` (standard step for docker swarm)
 - Set `gateway_public_url` as the dns domain pointing to the IP address of machine
 - Set `repository_url` and `push_repository_url` as "docker.io/{dockerhub_username}/"
 - Set `s3_url` to "minio:9000"

7. `faas-cli deploy -f stack.armhf.yml` to deploy all the main functions.

8. Install `auth` for github oauth:
   1. Create public and private keys for jwt token:
      ```bash
      # Private key
      openssl ecparam -genkey -name prime256v1 -noout -out key

      # Public key
      openssl ec -in key -pubout -out key.pub
      ```
   2. Store github app client secret:
      ```bash
      CLIENT_SECRET = "github app client secret"
      echo -n "${CLIENT_SECRET}" | docker secret create of-client-secret -
      ```
   3. Run auth service:
      ```bash
      docker service create --name auth \
      -e oauth_client_secret_path="/run/secrets/of-client-secret" \
      -e client_id="$CLIENT_ID" \
      -e PORT=8080 \
      -p 8085:8080 \
      -e external_redirect_domain="http://auth.system.subdomain:8081/" \
      -e cookie_root_domain=".system.subdomain" \
      -e public_key_path=/run/secrets/jwt-public-key \
      -e private_key_path=/run/secrets/jwt-private-key \
      -e oauth_provider="github" \
      --secret jwt-private-key \
      --secret jwt-public-key \
      --secret of-client-secret \
      --network func_functions \
      zeerorg/cloud-auth:armhf
      ```
    4. In the next step setup router to get github oauth url which can be set. That is, the oauth setup is complete after you deploy router.
    
9. Install `router` for doamin based URL
   1. Run command
      ```bash
      docker service rm of-router

      docker service create --network=func_functions \
       --env upstream_url=http://gateway:8080 \
       --env auth_url=http://auth:8080 \
       --publish 8081:8080 \
       --name of-router \
       -d zeerorg/cloud-router:armhf
       ```
   2. Your main url is now exposed at port http://subdomain:8081/
   3. Setup your "User authorization callback url" in the github app page to: "http://auth.system.subdomain:8081/"

10. Dashboard Install:
    1. Go into `/dashboard/` directory. Edit `dashboard_config.yml`
       - Remove `.openfaas` from `gateway_url`
       - Set `public_url` to `http://system.subdomain:8081/`
       - Set `pretty_url` to `http://user.subdomain:8081/function`
       - Set `base_href` to '/dashboard/'
    2. Deploy dashboard with:
       ```bash
       faas-cli template pull https://github.com/openfaas-incubator/node8-express-template
       faas-cli deploy --filter "system-dashboard" -f ./stack.armhf.yml
       ```


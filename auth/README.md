auth
=======

The auth service can be used to evaluate whether access is available to a given resource

For example:

Check access to resource (r) /function/system-dashboard:

```
http://auth:8080/q/?r=/function/system-dashboard
```

Responses:

* 200 - OK
* 301 - Cookie not present, redirect to given URL to create a valid cookie/login
* 401 - Cookie present, but invalid

Cookies:

* `openfaas_cloud`

This cookie is issued as part of the social sign-in flow using GitHub.

Contents (encoded JWT):

```json
{
  "name": "Alex Ellis",
  "access_token": "token-value",
  "aud": ".system.gw.io",
  "exp": 1537957152,
  "jti": "integer-value-here",
  "iat": 1537784352,
  "iss": "openfaas-cloud@github",
  "sub": "alexellis"
}
```

## Building

```
export TAG=0.3.0
make
```

## Running

All environmental variables must be set and configured for the service whether running locally as a container, via Swarm or on Kubernetes.


### Generate a key/pair

This key/pair is used to sign the JWT and then verify it later.

```
# Private key
openssl ecparam -genkey -name prime256v1 -noout -out key

# Public key
openssl ec -in key -pubout -out key.pub
```

For Kubernetes store these secrets:

```sh
kubectl -n openfaas create secret generic jwt-private-key --from-file=./key
kubectl -n openfaas create secret generic jwt-public-key --from-file=./key.pub
```

For Swarm you can create these secrets:

```sh
docker secret create jwt-private-key ./key
docker secret create jwt-public-key ./key.pub
```

### Store your `client_secret` in a secret


```sh
export CLIENT_SECRET=""
```

For Kubernetes store these secrets:


```sh
kubectl -n openfaas create secret generic of-client-secret --from-literal="of-client-secret=$CLIENT_SECRET"
```

For Swarm you can create these secrets:

```sh
echo -n "$CLIENT_SECRET" | docker secret create of-client-secret -
```

### As a local container:

```sh
docker rm -f cloud-auth
export TAG=0.3.0

docker run \
 -e oauth_client_secret_path="/run/secrets/of-client-secret" \
 -e client_id="$CLIENT_ID" \
 -e PORT=8080 \
 -p 8880:8080 \
 -e external_redirect_domain="http://auth.system.gw.io:8081" \
 -e cookie_root_domain=".system.gw.io" \
 -e public_key_path=/tmp/key.pub \
 -e private_key_path=/tmp/key \
 -v "`pwd`/key:/tmp/key" \
 -v "`pwd`/key.pub:/tmp/key.pub" \
 --name cloud-auth  -ti openfaas/cloud-auth:${TAG}
```

### On Kubernetes

Edit `yaml/core/of-auth-dep.yml` as needed and apply that file.

### On Swarm:

```sh
export TAG=0.3.0
docker service rm auth
docker service create --name auth \
 -e oauth_client_secret_path="/run/secrets/of-client-secret" \
 -e client_id="$CLIENT_ID" \
 -e PORT=8080 \
 -p 8085:8080 \
 -e external_redirect_domain="http://auth.system.gw.io:8081" \
 -e cookie_root_domain=".system.gw.io" \
 -e public_key_path=/run/secrets/jwt-public-key \
 -e private_key_path=/run/secrets/jwt-private-key \
 --secret jwt-private-key \
 --secret jwt-public-key \
 openfaas/cloud-auth:$TAG
```

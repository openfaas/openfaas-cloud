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

Contents:

(JSON payload)

| Field      | Description                     | Required |
|------------|---------------------------------|----------|
| sub        | unique ID for user              | true     |
| login      | login in GitHub                 | true     |
| full_name  | user's full-name as per profile | false    |

## Generate a key/pair

```
# Private key
openssl ecparam -genkey -name prime256v1 -noout -out key

# Public key
openssl ec -in key -pubout -out key.pub
```

For Kubernetes store these are secrets:

```sh
kubectl -n openfaas create secret generic jwt-private-key --from-file=./key
kubectl -n openfaas create secret generic jwt-public-key --from-file=./key.pub
```

## Building

```
export TAG=0.1.0
make
```

## Running 

```sh
docker rm -f cloud-auth
export TAG=0.1.0

docker run -e client_secret=x \
 -e client_id=y \
 -e PORT=8080 \
 -p 8080:8080 \
 -e external_redirect_domain="http://auth.system.gw.io:8081" \
 -e cookie_root_domain=".system.gw.io" \
 -e public_key_path=/tmp/key.pub \
 -e private_key_path=/tmp/key \
 -v `pwd`/key:/tmp/key \
 -v `pwd`/key.pub:/tmp/key.pub \
 --name cloud-auth  -ti openfaas/cloud-auth:$TAG
```

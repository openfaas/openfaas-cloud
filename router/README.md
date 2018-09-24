## Router for wildcard domain-name

This is a Golang reverse proxy which applies some mapping rules to let a user's wildcard domain name map back to a function route on the API gateway.

### Example:

Repo: alexellis
Function: kubecon-tester

Deployed function: alexellis-kubecon-tester

Gateway address: http://gateway:8080/function/alexellis-kubecon-tester

User-facing proxy address: https://alexellis.domain.io/kubecon-tester


### Usage:

```sh
upstream_url=http://gateway:8080 port=8081 go run main.go
```

Test it:

```sh
curl -H "Host: alexellis.domain.io" localhost:8081/kubecon-tester
```

### Development

```sh
TAG=0.5.0 make build ; make push
```

If you wish to bypass authentication you can run the router auth an auth_url of the `echo` function deployed via the `stack.yml`.

``` sh
TAG=0.5.0

# As a container

docker rm -f of-router
docker service rm of-router
docker run \
 -e upstream_url=http://127.0.0.1:8080 \
 -e auth_url=http://echo:8080 \
 -p 8081:8080 \
 --network=func_functions \
 --name of-router \
 -d openfaas/cloud-router:$TAG

# Or as a service

docker rm -f of-router
docker service rm of-router
docker service create --network=func_functions \
 --env upstream_url=http://127.0.0.1:8080 \
 --env auth_url=http://echo:8080 \
 --publish 8081:8080 \
 --name of-router \
 -d openfaas/cloud-router:$TAG
```

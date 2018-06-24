## Router for wildcard domain-name

This is a Golang reverse proxy which applies some mapping rules to let a user's wildcard domain name map back to a function route on the API gateway.

### Example:

Repo: alexellis
Function: kubecon-tester

Deployed function: alexellis-kubecon-tester

Gateway address: http://gateway:8080/function/alexellis-kubecon-tester

User-facing proxy address: https://alexellis.domain.io/kubecon-tester


### Usage:

```
upstream_url=http://gateway:8080 port=8081 go run main.go
```

Test it:

```
curl -H "Host: alexellis.domain.io" localhost:8081/kubecon-tester
```

### Development

```
TAG=0.1 make build ; make push 
```

```
TAG=0.1
docker rm -f of-router
docker service rm of-router
docker run -e upstream_url=http://192.168.0.35:8080 -p 8082:8080 --name of-router -d openfaas/cloud-router:$TAG

# Or as a service

TAG=0.1
docker rm -f of-router
docker service rm of-router
docker service create --env upstream_url=http://192.168.0.35:8080 --publish 8082:8080 --name of-router -d openfaas/cloud-router:$TAG
```
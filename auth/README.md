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


## Building

```
export TAG=0.1.0
make
```

## Running 

```
export TAG=0.1.0

docker rm -f cloud-auth ; \
docker run \
 -e PORT=8080 \
 -p 8080:8080 \
 -e client_secret=$CLIENT_SECRET \
 -e client_id=$CLIENT_ID \
 -e external_redirect_domain="http://auth.system.gw.io:8081" \
 -e cookie_root_domain=".system.gw.io"
 --name cloud-auth \
 -ti openfaas/cloud-auth:$TAG
```

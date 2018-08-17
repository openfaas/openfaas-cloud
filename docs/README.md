## OpenFaaS Cloud installation guide

## Intro

For the legacy instructions see [./README_LEGACY.md](./README_LEGACY.md)

### Pre-reqs

* Kubernetes or Docker Swarm
* Registry account - Docker Hub account or private registry with TLS
* OpenFaaS deployed with authentication enabled
* Extended timeouts for the queue-worker, gateway and the backend provider

For the queue-worker, set the `ack_wait` field to `15m` to allow for a Docker build of up to 15 minutes total.

### Moving parts

You will create/deploy:

* A GitHub App
* A number of secrets
* A stack of OpenFaaS functions via stack.yml
* Customise limits for Swarm or K8s
* Setup a container image builder

## Steps

### Create your GitHub App

Create a GitHub app from your GitHub profile page.

Select these OAuth permissions: 

- "Repository contents" read-only
- "Commit statuses" read and write

* Now select only the "push" event.

* Where can this GitHub App be installed?

Any account

* Save the new app.

* Now download the private key for your GitHub App which we will use later in the guide for allowing OpenFaaS Cloud to write to commit statuses when a build passes or fails.

The GitHub app will deliver webhooks to your OpenFaaS Cloud instance every time code is pushed in a user's function repository. Make sure you provide the public URL for your OpenFaaS gateway to the GitHub app. Like:  
`http://my.openfaas.cloud/function/github-push`

### Set your GitHub App config

#### Set the App ID

* Edit `github.yml` and populate with your GitHub App ID:

```yaml
environment:
    github_app_id: "<app_id>"
```

* Create a secret for the HMAC / webhook value:

Kubernetes:

```bash
WEBHOOK_SECRET="Long-Password-Phrase-Goes-Here"

kubectl create secret generic -n openfaas-fn github-webhook-secret --from-literal github-webhook-secret="$WEBHOOK_SECRET"
```

Swarm:

```bash
WEBHOOK_SECRET="Long-Password-Phrase-Goes-Here"
echo -n "$WEBHOOK_SECRET" | docker secret create github-webhook-secret -
```

#### Create a secret for your GitHub App's private key

Download the `.pem` file from the GitHub App page, then save it as a file named `private-key` with no extension.

* Kubernetes

```
kubectl secret create generic private-key --from-file=private-key -n openfaas-fn
```

* Docker Swarm

```
docker secret create private-key private-key
```

> Note: The default private key secret name is `private-key`. If needed different name can be specified by setting `private_key_filename` value in `github.yml`

```yaml
private_key_filename: my-private-key
```

### Customize for Kubernetes or Swarm

By default all settings are prepared for Kubernetes, so if you're using Swarm do the following:

#### Edit hostnames

In `gateway_config.yml` and `./dashboard/stack.yml` remove the suffix `.openfaas` where you see it.

#### Set limits

You will need to edit `stack.yml` and make sure `buildshiprun_limits_swarm.yml` is listed instead of `buildshiprun_limits_k8s.yml`.

### Deploy your container builder

Make sure Docker for Mac / Windows isn't storing your credentials in its keychain. If you're on Linux there is no need to make a change.

Log into your registry or the Docker hub

```
docker login
```

This populates `~/.docker/config.json` which is used in the builder.

#### For Kubernetes

Create the secret for your registry

```
kubectl create secret generic \
  --namespace openfaas \
  registry-secret --from-file=$HOME/.docker/config.json 
```

Create of-builder and of-buildkit:

```
kubectl apply -f ./yaml
```

#### For Swarm

Create the secret for your registry

```
cat $HOME/.docker/config.json | docker secret create registry-secret -
```

Create of-builder and of-buildkit:

```
./of-builder/deploy_swarm.sh
```

### Configure push repository and gateway URL

In gateway.yml

```yaml
environment:
  gateway_url: http://gateway.openfaas:8080/
  gateway_public_url: http://of-cloud.public-facing-url.com:8080/
  audit_url: http://gateway.openfaas:8080/function/audit-event
  repository_url: docker.io/ofcommunity/
  push_repository_url: docker.io/ofcommunity/
```

Replace "ofcommunity" with your Docker Hub account i.e. `alexellis2/cloud/` or replace the whole string with the address of your private registry `reg.my-domain.xyz`.

Now set your gateway's public URL in the `gateway_public_url` field.

### Configure pull secret

This is only needed if your registry uses authentication to pull images. The Docker Hub allows image to be pulled without a `pull secret`.

#### Swarm

* Uncomment the `registry-auth` secret in stack.yml

* Create the `registry-auth` pull secret:

Use the username and password from `docker login`

```
echo -n username:password | base64 | docker secret create registry-auth -
```

#### Kubernetes

Configure your service account with a pull secret as per [OpenFaaS docs](https://docs.openfaas.com/deployment/kubernetes/#link-the-image-pull-secret-to-a-namespace-service-account). Pick the `openfaas-fn` namespace.

### Create basic-auth secrets for the functions

The functions will need to use basic authentication to access the gateway's administrative endpoints.

Use the credentials you got when you set up OpenFaaS.

#### Swarm

```
echo "admin" > basic-auth-user
echo "admin" > basic-auth-password
docker secret create basic-auth-user basic-auth-user
docker secret create basic-auth-password basic-auth-password
```

#### Kubernetes

Create secrets in the openfaas-fn namespace:

```
echo "username" > basic-auth-user
echo "password" > basic-auth-password
kubectl create secret generic basic-auth-user --from-file=basic-auth-user=./basic-auth-user -n openfaas-fn
kubectl create secret generic basic-auth-password --from-file=basic-auth-password=./basic-auth-password -n openfaas-fn
```

### Deploy the OpenFaaS Cloud Functions

Optionally set the gateway URL:

```
export OPENFAAS_URL=https://gw.my-site.xyz
```

Now deploy:

```
faas-cli deploy
```

### Test it out

Now find the public URL of your GitHub App and navigate to it. Click "Install" and pick a GitHub repo you want to use.

Now create a new function, rename the YAML file for the functions to `stack.yml` and then commit it. When you put it up you'll see the logs in the `github-push` function.

#### Troubleshoot Kubernetes

Find out whether the pull/checkout/tar and build and deploy operation passed for each function.

```
kubectl logs -f deploy/github-push -n openfaas-fn
```

```
kubectl logs -f deploy/git-tar -n openfaas-fn
```

Find out if the build and deployment passed:

```
kubectl logs -f deploy/buildshiprun -n openfaas-fn
```

Find all events on the functions namespace

```
kubectl get events --sort-by=.metadata.creationTimestamp -n openfaas-fn
```

#### Troubleshoot Swarm

```
docker service logs github-push --tail 50
```

```
docker service logs git-tar --tail 50
```

```
docker service logs buildshiprun --tail 50
```

## Appendix

### Dashboard

The Dashboard is optional and can be installed to visualise your functions.

A pretty URL means that users get their own sub-domain per function. You will need to setup a wildcard DNS entry for this, but it's not essential, using the gateway address will also work.

Pretty vs non-pretty:

https://alexellis.domain.com/function1

https://gateway.domain.com/alexellis-function1

* On Docker Swarm you must remove the suffix `.openfaas` from the gateway address.

* Edit stack.yml

Set `query_pretty_url`  to `true` when using a sub-domain for each user, if not, then set this to an empty string or remove the line. If set, also define `pretty_url` with the pattern for the URL.

For a pretty URL you should also prefix each function with `system-` before deployinh.

Example with domain `o6s.io`:

```
      pretty_url: "http://user.o6s.io/function"
```

Set `public_url` to be the URL for the IP / DNS if not using a `pretty_url`

* Deploy

```
cd dashboard
faas-cli deploy
```

### SealedSecret support

The support for SealedSecrets is optional.

* Add the CRD entry for SealedSecret:

```sh
release=$(curl --silent "https://api.github.com/repos/bitnami-labs/sealed-secrets/releases/latest" | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p')
kubectl create -f https://github.com/bitnami-labs/sealed-secrets/releases/download/$release/sealedsecret-crd.yaml
```

* Install the CRD controller to manage SealedSecrets:

```sh
kubectl apply -f https://github.com/bitnami-labs/sealed-secrets/releases/download/$release/controller.yaml
```

* Install kubeseal CLI

You can perform the following two commands on the client or the server providing that you have a `.kube/config` file available and have switched to that context with `kubectl config set-context`.

```sh
release=$(curl --silent "https://api.github.com/repos/bitnami-labs/sealed-secrets/releases/latest" | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p')
GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)
wget https://github.com/bitnami/sealed-secrets/releases/download/$release/kubeseal-$GOOS-$GOARCH
sudo install -m 755 kubeseal-$GOOS-$GOARCH /usr/local/bin/kubeseal
```

* Export your public key

Now export the public key from Kubernetes cluster

```sh
kubeseal --fetch-cert > pub-cert.pem
```

You will need to distribute or share pub-cert.pem so that people can use this with the OpenFaaS CLI `faas-cli cloud seal` command to seal secrets.



### Wildcard domains with of-router

Coming soon
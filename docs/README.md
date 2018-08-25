## OpenFaaS Cloud installation guide

## Intro

For the legacy instructions see [./README_LEGACY.md](./README_LEGACY.md)

### Pre-reqs

* Kubernetes or Docker Swarm (Docker cannot be running with XFS as a backing file-system due to buildkit restrictions)
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
* (K8s only) [Network Policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/) for openfaas and openfaas-fn namespaces

### Before you begin

* You must enable basic auth to prevent user-functions from accessing the admin API of the gateway
* A list of valid users is defined in the CUSTOMERS file in this GitHub repo, this acts as an ACL, but you can define your own
* Swarm offers no isolation between functions (they can call each other)
* For Kubernetes isolation can be applied through [NetworkPolicy](https://kubernetes.io/docs/concepts/services-networking/network-policies/)

## Steps

### Create your GitHub App

Create a GitHub app from your GitHub profile page.

Select these OAuth permissions: 

- "Repository contents" read-only
- "Commit statuses" read and write
- "Checks" read and write

* Now select only the "push" event.

* Where can this GitHub App be installed?

Any account

* Save the new app.

* Now download the private key for your GitHub App which we will use later in the guide for allowing OpenFaaS Cloud to write to Checks or Commit statuses when a build passes or fails.

The GitHub app will deliver webhooks to your OpenFaaS Cloud instance every time code is pushed in a user's function repository. Make sure you provide the public URL for your OpenFaaS gateway to the GitHub app:

* With the router configured the URL should be like: `http://system.openfaas.cloud/github-event`

* If you don't have the router configured remove the `system-` prefix from `system-github-event` in `stack.yml` and set the URL like: `http://my.openfaas.cloud/functions/github-event`


### Create an internal trust secret

This secret will be used by each OpenFaaS Cloud function to validate requests and to sign calls it needs to make to other functions.

```
PAYLOAD_SECRET=$(head -c 12 /dev/urandom | shasum| cut -d' ' -f1)
```

Kubernetes:

```bash
kubectl create secret generic -n openfaas-fn payload-secret --from-literal payload-secret="$PAYLOAD_SECRET"
```

Swarm:

```bash
echo -n "$PAYLOAD_SECRET" | docker secret create payload-secret -
```

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
kubectl create secret generic private-key \
 --from-file=private-key -n openfaas-fn
```

* Docker Swarm

```
docker secret create private-key private-key
```

> Note: The default private key secret name is `private-key`. If needed different name can be specified by setting `private_key_filename` value in `github.yml`

```yaml
private_key_filename: my-private-key
```

### Setup your customer access control list (ACL)

Edit `customers_url` in gateway_config.yml.

Enter a list of GitHub usernames for your customers, these are case-sensitive.

### Customize for Kubernetes or Swarm

By default all settings are prepared for Kubernetes, so if you're using Swarm do the following:

#### Edit hostnames

In `gateway_config.yml` and `./dashboard/stack.yml` remove the suffix `.openfaas` where you see it.

#### Set limits

You will need to edit `stack.yml` and make sure `buildshiprun_limits_swarm.yml` is listed instead of `buildshiprun_limits_k8s.yml`.

### Deploy your container builder

You need to generate the ```~/.docker/config.json``` using the ```docker login``` command. 

If you are not on Linux, i.e. you are on Mac or Windows, docker stores credentials in credentials store by default and your docker config.json file will look like this:
```
{
  "credSstore" : "osxkeychain",
  "auths" : {
    "https://index.docker.io/v1/" : {

    }
  },
  "HttpHeaders" : {
    "User-Agent" : "Docker-Client/18.06.0-ce (darwin)"
  }
}
```
Run ```docker login``` to generate the ```config.json``` (if you haven't already) and edit it by removing the "credSstore" property:
```json
{
  "auths" : {
    "https://index.docker.io/v1/" : {

    }
  },
  "HttpHeaders" : {
    "User-Agent" : "Docker-Client/18.06.0-ce (darwin)"
  }
}
```
Log into your registry or the Docker hub

```
docker login
```
Expect to see ```WARNING! Your password will be stored unencrypted in /Users/kvuchkov/.docker/config.json``` in the output.

This populates `~/.docker/config.json` which is used in the builder:
```json
{
	"auths": {
		"https://index.docker.io/v1/": {
			"auth": "asdf12djs37ASfs732sFa3fdsw=="
		}
	},
	"HttpHeaders": {
		"User-Agent": "Docker-Client/18.06.0-ce (darwin)"
	}
}

```
> Note: the auth string above is using an example value and is not a real authentication string.

#### For Kubernetes

Create the secret for your registry

```
kubectl create secret generic \
  --namespace openfaas \
  registry-secret --from-file=$HOME/.docker/config.json 
```

Create of-builder, of-buildkit:

```
kubectl apply -f ./yaml/core
```

(Optional) Deploy NetworkPolicy
```
kubectl apply -f ./yaml/network-policy
```

(Optional) Add a role of "openfaas-system" using a label to the namespace where you deployed Ingress Controller. For example if Ingress Controller is deployed in the namespace `ingress-nginx`:
```
kubectl label namespace ingress-nginx role=openfaas-system
```

If you don't have Ingress Controller installed in cluster. [Read this](#troubleshoot-network-policies)

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

* Uncomment the `swarm-pull-secret` secret in stack.yml

* Create the `swarm-pull-secret` pull secret:

Use the username and password from `docker login`

```
echo -n username:password | base64 | docker secret create swarm-pull-secret -
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
kubectl create secret generic basic-auth-user \
 --from-file=basic-auth-user=./basic-auth-user -n openfaas-fn
kubectl create secret generic basic-auth-password \
 --from-file=basic-auth-password=./basic-auth-password -n openfaas-fn
```

### Log storage with Minio/S3

Logs from the container builder are stored in S3. This can be Minio which is S3-compatible, or AWS S3.

You can disable Log storage by commenting out the pipeline-log function from `stack.yml`.

* Generate secrets for Minio

```
SECRET_KEY=$(head -c 12 /dev/urandom | shasum| cut -d' ' -f1)
ACCESS_KEY=$(head -c 12 /dev/urandom | shasum| cut -d' ' -f1)
```

* If you'd prefer to use an S3 Bucket hosted on AWS

> Generate an access key by using the security credentials page. Expand the Access Keys section, and then Create New Root Key.

> Generate an secret key by opening the IAM console. Choose Users in the Details pane, pick the IAM user which will use the keys, and then Create Access Key on the Security Credentials tab.

```
SECRET_KEY=access_key_here
ACCESS_KEY=secret_key_here
```

#### Kubernetes

Store the secrets in Kubernetes

```
kubectl create secret generic -n openfaas-fn \
 s3-secret-key --from-literal s3-secret-key="$SECRET_KEY"
kubectl create secret generic -n openfaas-fn \
 s3-access-key --from-literal s3-access-key="$ACCESS_KEY"
```

Install Minio with helm

```
helm install --name cloud --namespace openfaas \
   --set accessKey=$ACCESS_KEY,secretKey=$SECRET_KEY,replicas=1,persistence.enabled=false,service.port=9000,service.type=NodePort \
  stable/minio
```

The name value should be `cloud-minio.openfaas.svc.cluster.local`

Enter the value of the DNS above into `s3_url` in `gateway_config.yml` adding the port at the end:`cloud-minio-svc.openfaas.svc.cluster.local:9000`

#### Swarm

Store the secrets

```
echo -n "$SECRET_KEY" | docker secret create s3-secret-key -
echo -n "$ACCESS_KEY" | docker secret create s3-access-key -
```

Deploy Minio

```
docker service rm minio

docker service create --constraint="node.role==manager" \
 --name minio \
 --detach=true --network func_functions \
 --secret s3-access-key \
 --secret s3-secret-key \
 --env MINIO_SECRET_KEY_FILE=s3-secret-key \
 --env MINIO_ACCESS_KEY_FILE=s3-access-key \
minio/minio:latest server /export
```

Minio: 

* Enter the value of the DNS above into `s3_url` in `gateway_config.yml` adding the port at the end:`minio:9000`

> Note: For debugging and testing. You can expose the port of Minio with `docker service update minio --publish-add 9000:9000`, but this is not recommended on the public internet.

AWS S3:

* Enter the value of the DNS `s3.amazonaws.com` into `s3_url` in `gateway_config.yml`

* In the same file set `s3_tls: true` and `s3_bucket` to the bucket you created in S3 like `s3_bucket: example-bucket`

* Update the `s3_region` value such as : `s3_region: us-east-1`


See https://docs.minio.io/docs/minio-quickstart-guide

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

##### Troubleshoot Network Policies
The NetworkPolicy configuration is designed to work with a Kubernetes IngressController. If you are using a NodePort or LoadBalancer you have to deploy NetworkPolicy below.
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: gateway
  namespace: openfaas
spec:
  policyTypes:
    - Ingress
  podSelector:
    matchLabels:
      app: gateway
  ingress:
    - from: []
      ports:
        - port: 8080
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

#### Prerequisites

The Dashboard is a SPA(Single Page App) made with React and will require the following:

- Node.js
- Yarn

#### Build and Bundle the Assets

If you have satisfied the prerequisites, the following command should create the assets for the Dashboard.

```bash
make
```

**Edit `stack.yml` if needed.**

Set `query_pretty_url` to `true` when using a sub-domain for each user. If set, also define `pretty_url` with the pattern for the URL.

Example with domain `o6s.io`:

```
pretty_url: "http://user.o6s.io/function"
```

Set `public_url` to be the URL for the IP / DNS of the OpenFaaS Cloud.

**Deploy**

> Don't forget to pull the `node8-express-template`

```
$ cd dashboard

$ faas-cli template pull https://github.com/openfaas-incubator/node8-express-template
$ faas-cli up
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

# OpenFaaS Cloud

## Conceptual diagram

This diagram shows the interactions between the functions that make up the OpenFaaS Cloud

![](https://pbs.twimg.com/media/DZiif9QXcAEd8If.jpg:large)

## Development

* Before you start

The public trial of OpenFaaS Cloud is running on Docker Swarm, so deploy OpenFaaS on Docker Swarm using [the documentation](https://docs.openfaas.com/deployment/).

You will also need to extend all timeouts for the gateway and the queue-worker. You should also increase the replica count of the `queue-worker` to enable concurrent builds for users.

The `ack_wait: "15m"` field should have a value of around `15m` to allow for a Docker build of up to 15 minutes.

OpenFaaS Cloud leverages OpenFaaS functions so will work with Kubernetes. Some minor tweaks may be needed to configuration YAML files such as URLs and memory limits for the of-builder and/or buildshiprun function. The shared cluster made available to the community runs on Kubernetes. 

> Note: it is easier to use one shared account on the Docker Hub or another public registry than to host your own registry with Kubernetes.

* Create secrets for the API Gateway

The API gateway uses secrets to enable basic authentication, even if your gateway doesn't have auth enabled you should create some temporary secrets or comment out the lines that bind secrets to functions in stack.yml.

Comment out these lines in stack.yml:

```yaml
      - basic-auth-user
      - basic-auth-password
```

Or create temporary secrets:

```sh
echo "admin" > basic-auth-user
echo "admin" > basic-auth-password
docker secret create basic-auth-user basic-auth-user
docker secret create basic-auth-password basic-auth-password
```

If you already have secrets set up for your gateway version 0.8.2 or later then there is no need to do the steps above.

* Create a GitHub app

Create a GitHub app from your GitHub profile page.

Before you start you'll need to create a free GitHub app and select these OAuth permissions: 

- "Repository contents" read-only
- "Commit statuses" read and write

* Now select only the "push" event.

* Where can this GitHub App be installed?

Any account

* Save the new app.

* Now download the private key for your GitHub App which we will use later in the guide for allowing OpenFaaS Cloud to write to commit statuses when a build passes or fails.

The GitHub app will deliver webhooks to your OpenFaaS Cloud instance every time code is pushed in a user's function repository. Make sure you provide the public URL for your OpenFaaS gateway to the GitHub app. Like:  
`http://my.openfaas.cloud/function/github-push`

* Create `github.yml` and populate it with your secrets as configured on the GitHub App:

```yaml
environment:
    github_app_id: "<app_id>"
```

A shared secret is used to securely verify each message came from GitHub and not a third party.

* Create a secret for the HMAC / webhook value:

Swarm:

```bash
WEBHOOK_SECRET="Long-Password-Goes-Here"
echo -n "$WEBHOOK_SECRET" | docker secret create github-webhook-secret -
```

Kubernetes:

```bash
WEBHOOK_SECRET="Long-Password-Goes-Here"

kubectl create secret generic -n openfaas-fn github-webhook-secret --from-literal github-webhook-secret="$WEBHOOK_SECRET"
```

### Status updates

OpenFaaS Cloud can update the statuses of your commits when a build passes or fails.

To enable this set `report_status: "true"` in `github.yml` before deploying the stack.

A private key must also be mounted as a Docker / Kubernetes secret so that the code can authenticate and update statuses.

* Create GitHub App secret

Download the .pem file from the GitHub App page, then save it as a file named `private-key` with no extension.

Create Docker secret

```
docker secret create private-key private-key
```

Default private key secret name is `private-key`
If needed different name can be specified by setting `private_key_filename` value in `github.yml`

```yaml
private_key_filename: my-private-key
```

* Update the remote gateway URL in `stack.yml` or set the `OPENFAAS_URL` environmental variable.

```yaml
provider:
  name: faas
  gateway: http://127.0.0.1:8080

```

You will need to edit `stack.yml` and make sure `buildshiprun_limits_swarm.yml` is listed instead of `buildshiprun_limits_k8s.yml`.

On Swarm you also need to edit all other YAML files to remove the `.openfaas` DNS suffix.

* Deploy of-builder and of-buildkit

Deploy the image builder (of-builder and of-buildkit) using these [instructions](./of-builder).

Optionally deploy a registry locally, or mount your `~/.docker/config.json` file to access a remote registry.

* Build/deploy (optional)

If you want to hack on the functions that make up OpenFaaS cloud then rebuild your own versions and push them to your Docker Hub account or another registry.

Before running this build/push/deploy script change the Docker Hub image prefix from `alexellis2/` to your own.

```sh
$ faas-cli build --parallel=4 \
  && faas-cli push --parallel=4 \
  && faas-cli deploy
```

* Test it out

Simply install your GitHub app to one of your OpenFaaS function repos and push a commit.

Within a few seconds you'll have your function deployed and live with a prefix of your username.

* Find out more

For more information get in touch directly for a private trial of the public service.

### UI Dashboard

The UI Dashboard shows your deployed functions by reading from the list-functions function. It is useful for testing and reviewing your functions as you go through a development workflow.

Deploy them separately from the [dashboard](/dashboard) folder.

### Appendix for Kubernetes

The functions which make up OpenFaaS Cloud are compatible with Kubernetes but some additional work is needed for the registry to make it work as seamlessly as it does on Swarm. The of-builder should also be brought up using a Kubernetes YAML file instead of `docker run` / `docker service create`.

You will need to edit `stack.yml` and make sure `buildshiprun_limits_k8s.yml` is listed instead of `buildshiprun_limits_swarm.yml`.

The way Kubernetes accesses the Docker registry will need a NodePort and a TLS cert (or an insecure-registry setting in the kubelet), so bear this in mind when working through the `of-builder` README.

* Update all timeouts

Before deploying you will need to bump all the timeouts to at least 300 seconds (5 minutes) to allow for latency building/pushing your images.

* Ensure basic auth is enabled on the gateway

`basic-auth-user` and `basic-auth-password` are to access gateway and are required when basic auth is enabled on the gateway service.

```yaml
      ....
      basic_auth: "true"
      ....
```

The functions will need basic auth credentials to access the gateway. Make sure they are defined in the cluster in the openfaas-fn namespace.

```yaml
      - basic-auth-user
      - basic-auth-password
```

Create basic auth secrets:

```sh
echo "username" > basic-auth-user
echo "password" > basic-auth-password
kubectl create secret generic basic-auth-user --from-file=basic-auth-user=./basic-auth-user -n openfaas-fn
kubectl create secret generic basic-auth-password --from-file=basic-auth-password=./basic-auth-password -n openfaas-fn
```

* Create a secret for the GitHub app

This should be the private key downloaded from GitHub

```
kubectl create secret generic private-key -n openfaas-fn --from-file=private-key=<path/to/private/key.pem>
```

* Edit github.yml

Set application ID

* Update `gateway_config.yml`

On Swarm you have to remove any suffixes found such as `.openfaas`.

I.e.

```yaml
environment:
  gateway_url: http://gateway.openfaas:8080/
  gateway_public_url: http://of-cloud.public-facing-url.com:8080/
  audit_url: http://gateway.openfaas:8080/function/audit-event
  repository_url: docker.io/ofcommunity/
  push_repository_url: docker.io/ofcommunity/
```

Replace "ofcommunity" with your Docker Hub account or the IP address of your registry.

* Update `buildshiprun`

Make sure you add the environment_file for `buildshiprun_limits_k8s.yml` and not for Swarm.

* Validate HMAC on `github-push`

Set `validate_hmac` to true and create a secret in the `openfaas-fn` namespace for `github-webhook-secret` which will be used to verify HMAC signatures of incoming webhooks. This is the webhook secret defined on your GitHub App.

* Deploy buildkit/of-builder

If you're using the Docker Hub then create a `config.json` file via `docker login` on Linux. This can then be set as a secret for buildkit in `of-builder-dep.yml`

```
$ kubectl create secret generic \
  --namespace openfaas \
  registry-secret --from-file=$HOME/.docker/config.json 
```

Update `of-builder-dep.yml` and the secret mount and volume projection to `/root/.docker/config.json`.

```yaml
    spec:
      volumes:
        - name: registry-secret
          secret:
            secretName: registry-secret
...

        securityContext:
          privileged: true
        volumeMounts:
        - name: quay-secret
          readOnly: true
          mountPath: "/root/.docker/"
```

You'll now need to deploy buildkit with `kubectl apply -f ./yaml/`

Delete the registry if you're using the Docker Hub: `kubectl delete -f ./yaml/registry.yml`

### Secrets

Secret support is available for functions through SealedSecrets, but you will need to install the Bitnami SealedSecrets controller before going any futher.

#### On the Server

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

* Seal a secret

To seal a secret type in:

```sh
faas-cli cloud seal --name alexellis-fn1 --literal secret="My AWS key goes here"
```

This will produce a secrets.yml file which can then be specified in your function definition as follows:

```yaml
    name: fn1
    secrets:
      - alexellis-fn1
```

Once ingested into the cluster via the `import-secrets` function you will see the following:

```sh
$ kubectl get sealedsecret -n openfaas-fn
NAME            AGE
alexellis-fn1   18s

$ kubectl get secret -n openfaas-fn
NAME                  TYPE                                  DATA      AGE
alexellis-fn1         Opaque                                1         16s
```

#### Troubleshooting on Kubernetes

Here are some commands to help you with troubleshooting your deployment on Kubernetes

Find out whether the pull/checkout/tar and build and deploy operation passed for each function.

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


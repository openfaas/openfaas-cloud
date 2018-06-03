# OpenFaaS Cloud

## Conceptual diagram

This diagram shows the interactions between the functions that make up the OpenFaaS Cloud

![](https://pbs.twimg.com/media/DZiif9QXcAEd8If.jpg:large)

## Development

* Before you start

The public trial of OpenFaaS Cloud is running on Docker Swarm, so deploy OpenFaaS on Docker Swarm using [the documentation](https://docs.openfaas.com/deployment/).

You will also need to extend all timeouts for the gateway and the queue-worker.

The `ack_wait: "300s"` field should have a value of around `300s` or 5 minutes to allow for a Docker build of up to 5 minutes.

OpenFaaS Cloud leverages OpenFaaS functions so will work with Kubernetes, but a complete configuration has not been provided yet. Some minor tweaks may be needed to configuration YAML files such as URLs and memory limits for the of-builder and/or buildshiprun function.

* Create secrets for the API Gateway

The API gateway uses secrets to enable basic authentication, even if your gateway doesn't have auth enabled you should create some temporary secrets or comment out the lines that bind secrets to functions in stack.yml.

Comment out these lines in stack.yml:

```
      - basic-auth-user
      - basic-auth-password
```

Or create temporary secrets:

```
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

Now select only the "push" event.

Now download the private key for your GitHub App which we will use later in the guide for allowing OpenFaaS Cloud to write to commit statuses when a build passes or fails.

The GitHub app will deliver webhooks to your OpenFaaS Cloud instance every time code is pushed in a user's function repository. Make sure you provide the public URL for your OpenFaaS gateway to the GitHub app. Like:  
`http://my.openfaas.cloud/function/gh-push`

* Create `github.yml` and populate it with your secrets as configured on the GitHub App:

```
environment:
    github_webhook_secret: "Long-Password-Goes-Here"
```
The shared secret is used to securely verify each message came from GitHub and not a third party.

* Add github appId to `github.yml`
```
environment:
    github_webhook_secret: "Long-Password-Goes-Here"
    github_app_id: "<app_id>"
```

### Status updates

OpenFaaS Cloud can update the statuses of your commits when a build passes or fails.

To enable this set `report_status: "true"` in `github.yml` before deploying the stack.

A private key must also be mounted as a Docker / Kubernetes secret so that the code can authenticate and update statuses.

* Find the .pem file from the GitHub App page

Create Docker secret
```
docker secret create derek-private-key <your_private_key_file>.pem
```

* Update the remote gateway URL in `stack.yml` or set the `OPENFAAS_URL` environmental variable.

```
provider:
  name: faas
  gateway: http://localhost:8080

```

* Deploy the registry and of-builder

Using the instructions given in the repo deploy of-builder (buildkit as a HTTP service) and the registry

https://github.com/openfaas/openfaas-cloud/tree/master/of-builder

* Build/deploy

> Before running this build/push/deploy script change the Docker Hub image prefix from `alexellis2/` to your own.

```
$ faas-cli build --parallel=4 \
  && faas-cli push --parallel=4 \
  && faas-cli deploy
```

* Test it out

Simply install your GitHub app to one of your OpenFaaS function repos and push a commit.

Within a few seconds you'll have your function deployed and live with a prefix of your username.

* Find out more

For more information get in touch directly for a private trial of the public service.

## UI Dashboard

The UI Dashboard shows your deployed functions by reading from the list-functions function. It is useful for testing and reviewing your functions as you go through a development workflow.

Deploy them separately from:

https://github.com/alexellis/of-cloud-fns

### Appendix for Kubernetes

The functions which make up OpenFaaS Cloud are compatible with Kubernetes but some additional work is needed for the registry to make it work as seamlessly as it does on Swarm. The of-builder should also be brought up using a Kubernetes YAML file instead of `docker run` / `docker service create`.

You will need to edit `buildshiprun_limits_swarm.yml` and change the unit value from `30m` to `30Mi` so that Kubernetes can make use of the value.

The way Kubernetes accesses the Docker registry will need a NodePort and a TLS cert (or an insecure-registry setting in the kubelet), so bear this in mind when working through the `of-builder` README.


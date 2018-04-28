# OpenFaaS Cloud

## Conceptual diagram

This diagram shows the interactions between the functions that make up the OpenFaaS Cloud

![](https://pbs.twimg.com/media/DZiif9QXcAEd8If.jpg:large)

## Development

* Before you start

The public trial of OpenFaaS Cloud is running on Docker Swarm, so deploy OpenFaaS on Docker Swarm using [the documentation](https://docs.openfaas.com/deployment/).

You will also need to extend all timeouts for the gateway and the queue-worker.

The `ack_timeout: "300s"` field should have a value of around `300s` or 5 minutes to allow for a Docker build of up to 5 minutes.

OpenFaaS Cloud leverages OpenFaaS functions so will work with Kubernetes, but a complete configuration has not been provided yet. Some minor tweaks may be needed to configuration YAML files such as URLs and memory limits for the of-builder and/or buildshiprun function.

* Create a GitHub app

Before you start you'll need to create a free GitHub app and select the relevant OAuth permissions. Right now those are just read-only and subscriptions to "push" events.

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

OpenFaaS Cloud
==============

OpenFaaS Cloud - FaaS in a box with CI/CD for functions

## Description

OpenFaaS Cloud uses serverless functions to provide a closed-loop CI/CD system for functions built and hosted on your public GitHub repositories.

OpenFaaS Cloud packages, builds and deploys functions using OpenFaaS functions written in Golang. Moby's BuildKit is used to build images and push to a local Docker registry instance.

Features:

* Applies GitOps principles - GitHub is the single source of truth
* To build and deploy a new version of a function - just push to your GitHub repo
* Subscription to OpenFaaS Cloud is done via a single click using a GitHub App
* Secured through HMAC - the public facing function "gh-push" uses HMAC to verify the origin of events

Conceptual diagram

![](https://pbs.twimg.com/media/DZ7SX6gX4AA5dS7.jpg:large)

## Functions

OpenFaaS Cloud is built using Golang functions to interact with GitHub and build/deploy your functions just seconds after your `git push`.

* Function: gh-push

Receives events from the GitHub app and checks the origin via HMAC

* Function: git-tar

Clones the git repo and checks out the SHA then uses the OpenFaaS CLI to shrinkwrap a tarball to be build with Docker

* Function: buildshiprun

Submits the tar to the of-builder then configures an OpenFaaS deployment based upon stack.yml found in the Git repo. Finally starts a rolling deployment of the function.

Calls garbage-collect

* Function: garbage-collect

Function cleans up functions which were removed or renamed within the repo for the given user.

* Service: of-builder

A builder daemon which exposes the GRPC of-buildkit service via HTTP.

* Service: of-buildkit

The buildkit GRPC daemon which builds the image and pushes it to the internal registry. The image is tagged with the SHA of the Git commit event.

* Service: Docker open-source registry

A private, local registry is deployed inside the cluster.

![](https://pbs.twimg.com/media/DZiif9QXcAEd8If.jpg:large)

## Usage

You can set up and host your own *OpenFaaS Cloud* or contact alex@openfaas.com for instructions on how to participate in a public trial of a fully-hosted service.

## Development

* Before you start

The public trial of OpenFaaS Cloud is running on Docker Swarm, so deploy OpenFaaS on Docker Swarm using [the documentation](https://docs.openfaas.com/deployment/).

OpenFaaS Cloud leverages OpenFaaS functions so will work with Kubernetes, but a complete configuration has not been provided yet. Some minor tweaks may be needed to configuration YAML files such as URLs and memory limits for the of-builder and/or buildshiprun function.

* Create a GitHub app

Before you start you'll need to create a free GitHub app and select the relevant OAuth permissions. Right now those are just read-only and subscriptions to "push" events.

The GitHub app will deliver webhooks to your OpenFaaS Cloud instance every time code is pushed in a user's function repository. Make sure you provide the public URL for your OpenFaaS gateway to the GitHub app.

* Create `github.yml` and populate it with your secrets as configured on the GitHub App:

```
environment:
    github_webhook_secret: "Long-Password-Goes-Here"
```

The shared secret is used to securely verify each message came from GitHub and not a third party.

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

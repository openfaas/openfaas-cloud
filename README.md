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

![](https://pbs.twimg.com/media/DZiif9QXcAEd8If.jpg:large)

Functions:

* gh-push

Receives events from the GitHub app and checks the origin via HMAC

* git-tar

Clones the git repo and checks out the SHA then uses the OpenFaaS CLI to shrinkwrap a tarball to be build with Docker

* buildshiprun

Submits the tar to the of-builder then configures an OpenFaaS deployment based upon stack.yml found in the Git repo. Finally starts a rolling deployment of the function.

Calls garbage-collect

* garbage-collect

Function cleans up functions which were removed or renamed within the repo for the given user.

* of-builder

A buildkit HTTP daemon which builds the image and pushes it to the internal registry. The image is tagged with the SHA of the Git commit event.

## Usage

You can set up and host your own *OpenFaaS Cloud* or contact alex@openfaas.com for instructions on how to participate in a public trial of a fully-hosted service.

## Development

Create `github.yml` and populate it with your secrets as configured on the GitHub App:

```
environment:
    github_webhook_secret: "Long-Password-Goes-Here"
```

Update the remote gateway URL in `stack.yml`

```
provider:
  name: faas
  gateway: http://localhost:8080

```

Build script:

```
$ faas-cli build -f stack.yml --parallel=4 \
  && faas-cli push -f stack.yml --parallel=4
  && faas-cli deploy -f stack.yml
```


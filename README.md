OpenFaaS Cloud
==============

OpenFaaS Cloud - FaaS in a box with CI/CD for your functions

![https://pbs.twimg.com/media/DacWCtZVMAAJQ-u.jpg](https://pbs.twimg.com/media/DacWCtZVMAAJQ-u.jpg)

*Announcement from Cisco's DevNet Create in Mountain View*

## Description

OpenFaaS Cloud uses serverless functions to provide a closed-loop CI/CD system for functions built and hosted on your public GitHub repositories. Just push your OpenFaaS functions to your public repo and within seconds you'll get a notificaiton with your HTTPS endpoint direcly on GitHub.

OpenFaaS Cloud packages, builds and deploys functions using OpenFaaS. Moby's BuildKit is used to build images and push to a local Docker registry instance.

Features:

* Applies GitOps principles - GitHub is the single source of truth
* To build and deploy a new version of a function - just push to your GitHub repo
* Subscription to OpenFaaS Cloud is done via a single click using a GitHub App
* Secured through HMAC - the public facing function "gh-push" uses HMAC to verify the origin of events
* HTTPS endpoint and build notifications for your commits

Conceptual diagram

![](https://pbs.twimg.com/media/DZ7SX6gX4AA5dS7.jpg:large)

## Functions

OpenFaaS Cloud is built using OpenFaaS Golang functions to interact with GitHub and build/deploy your functions just seconds after your `git push`.

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

## Try it out

You can set up and host your own *OpenFaaS Cloud* or contact alex@openfaas.com for instructions on how to participate in a public trial of a fully-hosted service.

Read the [development guide](DEV.md) to find out more about the functions and to start hacking on OpenFaaS Cloud.

OpenFaaS Cloud
==============

OpenFaaS Cloud - GitOps for your functions with native GitHub integrations

![https://pbs.twimg.com/media/DacWCtZVMAAJQ-u.jpg](https://pbs.twimg.com/media/DacWCtZVMAAJQ-u.jpg)

*Announcement from Cisco's DevNet Create in Mountain View*

## Description

OpenFaaS Cloud uses serverless functions to provide a closed-loop CI/CD system for functions built and hosted on your public GitHub repositories. Just push your OpenFaaS functions to your public repo and within seconds you'll get a notificaiton with your HTTPS endpoint direcly on GitHub.

OpenFaaS Cloud packages, builds and deploys functions using OpenFaaS. Moby's BuildKit is used to build images and push to a local Docker registry instance.

Features:

* Applies GitOps principles - GitHub is the single source of truth
* To build and deploy a new version of a function - just push to your GitHub repo
* Subscription to OpenFaaS Cloud is done via a single click using a GitHub App
* Secured through HMAC - the public facing function "github-event" uses HMAC to verify the origin of events
* HTTPS endpoint and build notifications for your commits

## Blog post

Read my [introducing OpenFaaS Cloud](https://blog.alexellis.io/introducing-openfaas-cloud/) blog post for an overview of the idea with examples, screenshots and background on the project.

## Roadmap & Features

* Core experience

- [x] Self-hosted deployment on Kuberneters or Docker Swarm
- [x] GitHub Status API integration for commits
- [x] Automatic HTTPS for endpoints (yes via CertManager/Traefik)
- [x] Authorization with a CUSTOMER list
- [x] Trust with GitHub via HMAC and GitHub App
- [x] CI/CD for functions
- [x] Container builder using BuildKit
- [x] Free to use SaaS edition for community members and contributors
- [x] Log storage on Minio (S3-compatible)
- [x] Log storage on AWS S3
- [ ] Kubernetes helm chart (plain YAML supported already)
- [ ] Automation for the day-0 installation via Ansible or similar

* Developer story

- [x] Multi-user
- [x] UI: [Dashboard for users](./dashboard)
- [x] Support secrets in public repos through Bitnami SealedSecrets
- [x] Make detailed logs available to show build or unit test failures (dashboard)
- [x] Make build logs available publicly (dashboard finished, Checks API in progress)
- [x] Mixed-case user-names
- [ ] Use a git "tag" or "GitHub release" to promote a function to live
- [ ] UI: Dashboard - detailed metrics of success/failure per function in

* Operationalize

- [x] Support for shared Docker Hub accounts instead of private registry
- [x] Support for private GitHub repos
- [x] Dashboard: OAuth 2 login via GitHub
- [x] Isolation between functions (through the provided NetworkPolicy on Kubernetes)

* Stretch goals

- [x] Move Dashboard UI to React.js
- [ ] Re-write React.js Dashboard to use native Bootstrap library
- [ ] CI/CD integration with on-prem GitLab (in-progress)
- [ ] UI: OAuth 2 login via GitLab (planned, help wanted)
- [ ] Unprivileged builds with BuildKit or similar (under investigation)
- [ ] Log into OpenFaaS Cloud via CLI (faas-cli cloud login)
- [ ] Enable untrusted container builds via docker-machine?
- [ ] Integration with on-prem BitBucket (help wanted)

## Functions

OpenFaaS Cloud is built using OpenFaaS Golang functions to interact with GitHub and build/deploy your functions just seconds after your `git push`.

* Function: github-event

Receives events from the GitHub app and checks the origin via HMAC

* Function: github-push

Handles push events from the "github-event" function

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

## Try it out

![](https://pbs.twimg.com/media/DZ7SX6gX4AA5dS7.jpg:large)

*Conceptual diagram of how OpenFaaS Cloud integrates with GitHub*

You can set up and host your own *OpenFaaS Cloud* or contact alex@openfaas.com for instructions on how to participate in a public trial of a fully-hosted service. Read the privacy statement and terms and conditions for the hosted version of [OpenFaaS Cloud](./PRIVACY.md).

Read the [development guide](docs/README.md) to find out more about the functions and to start hacking on OpenFaaS Cloud.

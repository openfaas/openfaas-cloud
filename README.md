OpenFaaS Cloud
==============

OpenFaaS Cloud: multi-user serverless functions managed with git

![https://pbs.twimg.com/media/DacWCtZVMAAJQ-u.jpg](https://pbs.twimg.com/media/DacWCtZVMAAJQ-u.jpg)

*Announcement from Cisco's DevNet Create in Mountain View*

## Description

OpenFaaS Cloud introduces an automated build and management system for your Serverless functions with native integrations into your source-control management system whether that is GitHub or GitLab.

With OpenFaaS Cloud functions are managed through typing `git push` which reduces the tooling and learning curve required to operate functions for your team.

As soon as OpenFaaS Cloud receives a `push event` from `git` it will run through a build-workflow which clones your repo, builds a Docker image, pushes it to a registry and then deploys your functions to your cluster. Each user can access and monitor their functions through their personal dashboard.

Features:

* Portable - self-host or use the hosted Community Cluster (SaaS)
* Multi-user - use your GitHub/GitLab identity to log into your personal dashboard
* Automates CI/CD triggered by `git push` (also known as GitOps)
* Onboard new git repos with a single click by adding the *GitHub App* or a repository tag in *GitLab*
* Immediate feedback on your personal dashboard and through GitHub Checks or GitLab Statuses
* Sub-domain per user or organization with HTTPS
* Fast, non-root image builds using Docker's buildkit

The dashboard page for a user:

![Dashboard](https://user-images.githubusercontent.com/6358735/46193701-f56b6680-c2f6-11e8-8bf4-9256a8341960.png)

The details page for a function:

![Details page](https://user-images.githubusercontent.com/6358735/46193700-f56b6680-c2f6-11e8-9bec-40b61e42ce45.png)

### Requirements

* OpenFaaS (0.9.10 or greater is recommended)
* Docker Swarm or Kubernetes (other OpenFaaS providers may work in the future)

## Blog post

Read my [introducing OpenFaaS Cloud](https://blog.alexellis.io/introducing-openfaas-cloud/) blog post for an overview of the idea with examples, screenshots and background on the project.

## Conceptual architecture diagram

This conceptual diagram shows how OpenFaaS Cloud integrates with GitHub/GitLab through the use of an event-driven architecture.

Main flows:

1. User pushes code - GitHub/GitLab push event is sent to github-event/gitlab-event function triggering a CI/CD workflow
2. User removes GitHub/GitLab app from one or more repos - garbage collection is invoked removing 1-many functions
3. User accesses function via router using "pretty URL" format and request is routed to function via API Gateway

![](./docs/conceptual-overview.png)

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

- [x] Trust between functions using HMAC and a shared secret.
- [x] Support for shared Docker Hub accounts instead of private registry
- [x] Support for private GitHub repos
- [x] Dashboard: OAuth 2 login via GitHub
- [x] Isolation between functions (through the provided NetworkPolicy on Kubernetes)

* Stretch goals

- [x] Move Dashboard UI to React.js
- [x] Re-write React.js Dashboard to use native Bootstrap library
- [ ] CI/CD integration with on-prem GitLab (in-progress)
- [x] UI: OAuth 2 login via GitLab
- [ ] Unprivileged builds with BuildKit or similar (under investigation)
- [ ] Log into OpenFaaS Cloud via CLI (faas-cli cloud login)
- [ ] Enable untrusted container builds via docker-machine?
- [ ] Integration with on-prem BitBucket (help wanted)

## Components

OpenFaaS Cloud is built primarily using OpenFaaS functions written in Golang, a router microservice, a container builder microservice (buildkit) and an auth microservice for (OAuth 2.0).

### Microservices

* Microservice: of-builder

A builder daemon which exposes the GRPC of-buildkit service via HTTP.

* Microservice: of-buildkit

The buildkit GRPC daemon which builds the image and pushes it to the internal registry. The image is tagged with the SHA of the Git commit event.

* Microservice: of-router

The router component is the only ingress point for HTTP requests for serving functions and for enabling the GitHub/GitLab integration. It translates "pretty URLS" into URLs namespaced by a user prefix on the OpenFaaS API Gateway.

* Microservice: of-auth

The auth service validates routes, can issue a JWT token and is called by the router component for every HTTP request.

* Service: Docker open-source registry

A private, local registry is deployed inside the cluster.

### Functions

* Function: github-event

Receives events from the GitHub app and checks the origin via HMAC with a shared secret with GitHub

* Function: github-push

Handles push events from the "github-event" function

* Function: git-tar

Clones the git repo and checks out the SHA then uses the OpenFaaS CLI to shrinkwrap the function's code into a tarball to be built by buildkit into a Docker image.

* Function: import-secrets

Used only with Kubernetes when SealedSecrets are installed. Binds SealedSecrets into the cluster so that the `buildshiprun` function can bind (unsealed) user secrets to functions.

* Function: pipeline-log

Either writes a build log or fetches one from an S3 bucket

* Function: list-functions

When passed a querystring of `?username=owner` this queries the API Gateway's `/system/functions` endpoint and then filters the content for the user.

* Function: buildshiprun

Submits the tar to the of-builder then configures an OpenFaaS deployment based upon `stack.yml` found in the Git repo. A rolling update is then sent to the API Gateway using basic auth followed by calling garbage-collect to remove old or orphaned functions.

* Function: github-status

Writes statuses to GitHub Checks API showing build status and URLs for endpoints

* Function: garbage-collect

Removes functions which were removed or renamed within the repo for the given user. Also responsible for handling requests to uninstall GitHub/GitLab app from a repo or account.

* Function: audit-event

Collects events from other functions for auditing. These can be connected to a Slack webhook URL or the function can be swapped for the echo function for storage in container logs.

* Function: system-metrics

Handler folder should be renamed to just `metrics`. Function can provide stats on invocations for function over given time period split by success/error.

## Try it out

You can set up and host your own *OpenFaaS Cloud* or contact alex@openfaas.com for instructions on how to participate in a public trial of a fully-hosted service (a.k.a. Community Cluster). Read the privacy statement and terms and conditions for the hosted version of [OpenFaaS Cloud](./PRIVACY.md).

Read the [development guide](docs/README.md) to find out more about the functions and to start hacking on OpenFaaS Cloud.

## Getting help

For help join #openfaas-cloud on the [OpenFaaS Slack workspace](https://docs.openfaas.com/community).

## OpenFaaS Cloud Components

OpenFaaS Cloud is built primarily using OpenFaaS functions written in Golang, a router microservice, a container builder microservice (buildkit) and an auth microservice for (OAuth 2.0).

### Conceptual architecture diagram

This conceptual diagram shows how OpenFaaS Cloud integrates with GitHub/GitLab through the use of an event-driven architecture.

Main flows:

1. User pushes code - GitHub/GitLab push event is sent to github-event/gitlab-event function triggering a CI/CD workflow
2. User removes GitHub/GitLab app from one or more repos - garbage collection is invoked removing 1-many functions
3. User accesses function via router using "pretty URL" format and request is routed to function via API Gateway

![](./docs/conceptual-overview.png)


### Microservices

* Microservice: of-builder

A builder daemon which exposes the GRPC of-buildkit service via HTTP.

* Microservice: of-buildkit

The buildkit GRPC daemon which builds the image and pushes it to the internal registry. The image is tagged with the SHA of the Git commit event.

* Microservice: edge-router

The router component is the only ingress point for HTTP requests for serving functions and for enabling the GitHub/GitLab integration. It translates "pretty URLS" into URLs namespaced by a user prefix on the OpenFaaS API Gateway.

* Microservice: edge-auth

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

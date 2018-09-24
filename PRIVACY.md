## OpenFaaS Cloud (hosted version)

This statement applies to the public, hosted version of OpenFaaS Cloud operated by Alex Ellis. When running OpenFaaS Cloud on your own OpenFaaS cluster these terms do not apply.

### Terms of use

In order to make use of OpenFaaS Cloud you need to consent to add a GitHub App to each of your GitHub repositories, you also need to add yourself to the CUSTOMERS file in this repository which maintains a list of all people participating in the public hosted environment. By adding your GitHub user account to the CUSTOMERS file you are giving consent for OpenFaaS Cloud to clone, build, and deploy your functions and make them available via a public HTTP endpoint.

The system is monitored and will be subjected to fair use policies.

The operator of OpenFaaS Cloud reserves the right to remove or revoke individual accounts from the public trial or to end the trial and remove the hosted service.

Please see the [LICENSE](./LICENSE.md) of the project for further details.

## Privacy policy

After having given consent for OpenFaaS Cloud to integrate with one or more of your GitHub repositories the following information will be recorded:

Within the OpenFaaS API:

* GitHub username
* GitHub repository name
* Container builder logs from the of-builder component for each function triggered for a build via "git push"

Stored within the local Kubernetes kubelet or Docker Swarm node within container log files:

* GitHub username
* GitHub repository name

Your public GitHub repository will be cloned and built, if successful the resulting Docker image will be stored in a Docker registry. Your function will be deployed and a public endpoint will be exposed to access your function.

No secret information from your GitHub profile is stored in OpenFaaS Cloud and minimal access is granted through OAuth permission scopes including: read/write to statuses and reading repository contents.

### Authentication

When the feature is live OpenFaaS Cloud will require a login to access the UI dashboard. Access to this dashboard will be via single sign-on (SSO) with your GitHub account through use of a GitHub OAuth 2.0 App. Once authenticated a JWT token will be issued to you containing claims about your user account - the JWT is stored in a Cookie in your browser, but is not stored on  OpenFaaS Cloud. If you wish to remove your cookie, just remove it from your browser or clear your browsing history.

You may retract access to the GitHub OAuth 2.0 App through your GitHub profile.

### Uninstalling

If you want to remove your endpoints from OpenFaaS Cloud please remove the GitHub App integration and GitHub OAuth 2.0 App via your GitHub profile page. This action will cause the deployed functions to be deleted, but the Docker images may still remain in the registry for a period of time after this step depending on whether a public Docker Hub account is being used or a private registry. To have all information removed contact alex@openfaas.com.


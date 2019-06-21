## OpenFaaS Cloud - Community Cluster (hosted version)

This statement applies to the public, hosted version of OpenFaaS Cloud (aka the Community Cluster) which is operated by OpenFaaS Ltd, registered company no: 11076587.

> Note: These terms do not apply if you are running your own OpenFaaS Cloud cluster.

### Terms of use

In order to make use of OpenFaaS Cloud you need to consent to add a GitHub App to each of your GitHub repositories, you also need to add yourself to the CUSTOMERS file in this repository which maintains a list of all people participating in the public hosted environment. By adding your GitHub user account to the CUSTOMERS file you are giving consent for OpenFaaS Cloud to clone, build, and deploy your functions and make them available via a public HTTP endpoint.

#### Acceptable use

The system is monitored and will be subjected to fair use policies.

You agree that you will:

* Not misuse the cluster
* Not attempt to gain access to other users' data or system services
* Not carry out unauthorized load tests or Denial of Service attacks
* Not publish statistics, reviews or social-media commentary about the cluster

If you wish to do any of the above, then [deploy your own environment](https://github.com/openfaas-incubator/ofc-bootstrap/) where these terms and conditions would not apply.

##### To report a vulnerability

Do not report a vulnerability via Slack or GitHub. Please contact OpenFaaS Ltd via alex@openfaas.com directly. Give 5-10 working days for an initial response.

##### Termination of access

The operator of OpenFaaS Cloud reserves the right to remove or revoke individual accounts from the public trial or to end the trial and remove the hosted service.

Disclaimer: You must agree to the terms of the project [LICENSE](./LICENSE.md).

## Privacy policy

After having given consent for OpenFaaS Cloud to integrate with one or more of your GitHub repositories the following information will be recorded:

Within the OpenFaaS API:

* GitHub username
* GitHub repository name
* Container builder logs from the of-builder component for each function triggered for a build via "git push"
* Any SealedSecrets you upload through your git repository

Stored within the local Kubernetes `kubelet` within container log files:

* GitHub username
* GitHub repository name

Your public GitHub repository will be cloned and built, if successful the resulting Docker image will be stored in a Docker registry. Your function will be deployed and a public endpoint will be exposed to access your function.

No secret information from your GitHub profile is stored in OpenFaaS Cloud and minimal access is granted through OAuth permission scopes including: read/write to statuses and reading repository contents.

### Authentication

When a user attempts to access their dashboard, OpenFaaS Cloud will require authentication.

Access to this dashboard will be via single sign-on (SSO) with your GitHub account through use of a GitHub OAuth 2.0 App. Once authenticated a JWT token will be issued to you containing claims about your user account - the JWT is stored in a Cookie in your browser, but is not stored on  OpenFaaS Cloud. If you wish to remove your cookie, just remove it from your browser or clear your browsing history.

You may retract access to the GitHub OAuth 2.0 App through your GitHub profile.

### Uninstalling

If you want to remove your endpoints from OpenFaaS Cloud please remove the GitHub App integration and GitHub OAuth 2.0 App via your GitHub profile page. This action will cause the deployed functions to be deleted, but the Docker images may still remain in the registry for a period of time after this step depending on whether a public Docker Hub account is being used or a private registry. 

#### Data protection

The "Community Cluster" is hosted on servers owned and controlled by [DigitalOcean LCC](https://www.digitalocean.com/about/) in the the London region. The operator of the "Community Cluster" has no affiliation with DigitalOcean.

To request your data or to have have all information removed contact OpenFaaS Ltd via alex@openfaas.com.


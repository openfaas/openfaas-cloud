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

Within container log files:

* GitHub username
* GitHub repository name

Your public GitHub repo will be cloned and built, if successful the resulting Docker image will be stored in a Docker registry. Your function will be deployed and a public endpoint will be exposed to access your function.

No secret information is stored in OpenFaaS Cloud and minimal access is granted through OAuth permission scopes including: read/write to statuses and reading repository contents.

Uninstalling

If you want to remove your endpoints from OpenFaaS Cloud please remove the GitHub app integration via your GitHub profile page and contact alex@openfaas.com to have your functions and container images removed from OpenFaaS Cloud.


## OpenFaaS Cloud

This is the experimental Helm Chart for installing OpenFaaS Cloud.

It is currently not recommended to use this chart directly, it is intended to be used as part of `ofc-bootstrap`
to install OpenFaaS Cloud clusters.

This is mainly for configuration documentation for developers wishing to use this chart in `ofc-bootstrap`.


## Configuration

| Parameter | Description | Default|
| --------- | ------- | ---------- |
| `ofBuilder.image` | `The image used to the OpenFaaS Builder` | `openfaas/of-builder:0.7.2` |
| `ofBuilder.replicas` | `The number of replicas for the OpenFaaS Builder deployment` | `1` |
| `buildKit.image` | `The Buildkit image used by OpenFaaS Cloud` | `moby/buildkit:v0.6.2` |
| `buildKit.privileged` | `If the buildKit container should run in privilaged mode` | `true` |
| `edgeAuth.image` | `The edge-auth image for OpenFaaS Cloud` | `openfaas/edge-auth:0.7.0` |
| `edgeAuth.replicas` | `Number of replicas of edge-auth to run` | `1` |
| `edgeAuth.enableOauth2` | `If OAuth2 is enabled in the installation` | `true` |
| `edgeAuth.oauthProvider` | `The OAuth provider, github or gitlab` | `github` |
| `edgeAuth.oauthProviderBaseURL` | `The OAuth2 base URL, required if using gitlab as OAuth2 provider` | `` |
| `edgeAuth.clientId` | `The client ID provided by your OAuth provider` | `""` |
| `edgeAuth.writeDebug` | `If Debug logging is enabled in edge-auth` | `false` |
| `edgeAuth.clientSecret` | `The client secret provided by your OAuth2 provider` | `""` |
| `edgeRouter.image` | `The container image for the OpenFaaS Cloud edge-router` | `openfaas/edge-router:0.7.4` |
| `tls.enabled` | `If we are using TLS, certificated provided by LEtsEncrypt` | `true` |
| `tls.email` | `The email for the LetsEncrypt TLS certificates. Required if TLS is enabled` | `example@example.com` |
| `tls.issuerType` | `The LetsEncrypt issuer to use, either staging or prod` | `staging` |
| `tls.dnsService` | `The DNS provider for your domain, one of route53, digitalocean, cloudflare or clouddns` | `digitalocean` |
| `tls.route53.region` | `The AWS Region for the DNS api calls` | `eu-west-1` |
| `tls.route53.accessKeyID` | `The Acces Key ID for route53, required if using route53 and ambientCredentials is false` | `` |
| `tls.route53.ambientCredentials` | `If true, this will instruct cert-manager to use an ambient credentials provider, if configured (such as kube2iam)` | `false` |
| `tls.cloudflare.email` | `If using Cloudflare for DNS set this to your Cloudflare email` | `` |
| `tls.clouddns.projectID` | `If using Clouddns set this to your project ID` | `` |
| `ingress.class` | `The ingress class used for ingress. Set to traefik if using traefik for ingress for example` | `nginx` |
| `ingress.maxConnections` | `The max connections allowed for OpenFaaS Cloud Functions` | `` |
| `ingress.requestsPerMinute` | `The max number of connections for ingress when using nginx for ingress class` | `20` |
| `ingesss.requestsPerMinute` | `The max requests per minute when using nginx for the inress class` | `600` |
| `customers.url` | `The public URL to a customers file, it should be unformatted with 1 username per line` | `""` |
| `customers.customersSecret` | `If set to ture we use a secret for our customers list rather than a public URL` | `false` |
| `global.rootDomain` | `The root domain for this OFC installation` | `"example.com"` |
| `global.scheme` | `Wither http or https, depending on the TLS setting` | `https` |
| `global.enableECR` | `Set to true is using ECR as our container registry rather than Docker Hub` | `false` |
| `global.imagePullPolicy` | `The policy for pulling OFC images, Allways or IfNotPresent for example` | `IfNotPresent` |
| `global.coreNamespace` | `The namespace for the core OFC components` | `openfaas` |
| `global.functionsNamespace` | `The namespace for the User Functions` | `openfaas-fn` |
| `global.httpProbe` | `The setting to detemine is we are using http probe or exec probe. Istio users may wish to set this to fale doe example.` | `true` |

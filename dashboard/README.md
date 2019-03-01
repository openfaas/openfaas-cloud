# OpenFaaS Cloud Dashboard

## Installation

The Dashboard is optional and can be installed to visualise your functions.

### Prerequisites

The Dashboard is a SPA(Single Page App) made with React and will require the following:

- faas-cli
- OpenFaaS deployed locally on Swarm or Kubernetes
- Docker
- Node.js LTS
- yarn

### Deploy a few functions for sample data

This may be simpler than deploying the builder and connecting your OpenFaaS Cloud to a GitHub App.

```
$ faas-cli store deploy figlet --name alexellis-figlet --label Git-Owner=alexellis \
 --label com.openfaas.cloud.git-repo=figlet \
 --label com.openfaas.cloud.git-sha=665d9597547d8e0425630ba2dbb73c2951a61ce2 \
 --label com.openfaas.cloud.git-deploytime=1533026741 \
 --label com.openfaas.cloud.git-cloud=1 --network=func_functions \
 --label com.openfaas.scale.min=1 --label com.openfaas.scale.max=4 \
 --annotation com.openfaas.cloud.git-repo-url=https://github.com/alexellis/figlet
```

```
$ faas-cli store deploy nodeinfo --name alexellis-nodeinfo \
 --label com.openfaas.cloud.git-owner=alexellis \
 --label com.openfaas.cloud.git-repo=nodeinfo \
 --label com.openfaas.cloud.git-sha=665d9597547d8e0425630ba2dbb73c2951a61ce2 \
 --label com.openfaas.cloud.git-deploytime=1533026741 \
 --label com.openfaas.cloud.git-cloud=1 \
 --label com.openfaas.scale.min=1 --label com.openfaas.scale.max=4 --network=func_functions \
 --annotation com.openfaas.cloud.git-repo-url=https://github.com/alexellis/nodeinfo
```

### Deploy at least the list-functions function

From the root directory edit `gateway_config.yml`, if on Swarm remove any `.openfaas` suffix you see in URLs.

Now run:

```
$ faas-cli deploy --filter="list-functions"
```

The `list-functions` function will be accessed by the dashboard and will be querying the metadata from the functions we deployed in the previous step.

### Edit stack.yml for the dashboard

From the dashboard folder make edits to `stack.yml` (read the comments in the file) for the dashboard, then deploy only the new dashboard function:

```
$ faas-cli deploy --filter="system-dashboard"
```

### Build and Bundle the Assets

If you have satisfied the prerequisites, the following command should create the assets for the Dashboard.

```bash
make
```

**Edit `stack.yml` if needed.**

Set `query_pretty_url` to `true` when using a sub-domain for each user. If set, also define `pretty_url` with the pattern for the URL.

Example with domain `o6s.io`:

```
pretty_url: "http://user.o6s.io/function"
```

Set `public_url` to be the URL for the IP / DNS of the OpenFaaS Cloud.

Set `cookie_root_domain` when using auth.

Example with domain `o6s.io`:

```
cookie_root_domain: ".system.o6s.io"
```

**Deploy**

> Don't forget to pull the `node10-express-template`

```
$ faas-cli template pull https://github.com/openfaas-incubator/node10-express-template
$ faas-cli up
```

## Development

Install `yarn` (requires Node.js LTS)

```
npm i -g yarn
```

The source code for the dashboard (written in React.js) with Bootstrap 3 has to be built into a generated folder. In order to do this type in `make`

```bash
make
```

You will see new files written into `of-cloud-dashboard/dist`

When you are ready to make a Pull Request, do not commit the `of-cloud-dashboard/dist` folder since it will cause issues with multiple developers clashing over the same files. Instead only commit the `client` folder and make sure that you ignore any edits that you have made to `stack.yml` for your local environment.

**Set Proxy**

Depending on your environment, set the proxy for the webpack devServer. By default, all `/api` requests get proxied to `http://127.0.0.1:8080/function` which will usually be your local OpenFaaS gateway. Change this inside the `package.json` for your environment.

```json
  "proxy": {
    "/api": {
      "pathRewrite": {
        "^/api": "/"
      },
      "target": "http://127.0.0.1:8080/function",
    }
  },
```

**Run Dev Server**

```bash
# Start the dev server
yarn start
```

You can connect to the UI at http://127.0.0.1:3000.

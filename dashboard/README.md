# OpenFaaS Cloud Dashboard

## Installation

The Dashboard is optional and can be installed to visualise your functions.

### Prerequisites

The Dashboard is a SPA(Single Page App) made with React and will require the following:

- Node.js
- Yarn

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

**Deploy**

```
$ faas-cli deploy
```

## Development

**Install Dependencies**

```bash
# Move to the client directory
cd client
# Install dependencies
yarn
```

**Set Proxy**

Depending on your environment, set the proxy for the webpack devServer. By default, all `/api` requests get proxied to `http://localhost:8080/function` which will usually be your local OpenFaaS gateway. Change this inside the `package.json` for your environment.

```json
  "proxy": {
    "/api": {
      "pathRewrite": {
        "^/api": "/"
      },
      "target": "http://localhost:8080/function",
    }
  },
```

**Run Dev Server**

```bash
# Start the dev server
yarn start
```

You can connect to the UI at http://localhost:3000.

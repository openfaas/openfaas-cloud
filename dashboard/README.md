OpenFaaS Cloud Dashboard 
==============================

This code originated from [functions built by Alex Ellis](https://github.com/alexellis/of-cloud-fns/blob/master/overview/handler.go) for exploring OpenFaaS Cloud visually.

## Installation

The Dashboard is optional and can be installed to visualise your functions.

* Edit stack.yml if needed.

Set `query_pretty_url`  to `true` when using a sub-domain for each user. If set, also define `pretty_url` with the pattern for the URL.

Set `public_url` to be the URL for the IP / DNS if not using a `pretty_url`

* Deploy

```
$ faas-cli deploy
```


## system-metrics

This function exposes metrics from Prometheus.

It queries Prometheus for a function's 2xx and non-2xx invocations count and returns a JSON response of type:

```json
{
    "success": 10,
    "failure": 8
}
```

It takes function's name (i.e. `myFunction` \[required\]) and metrics_window (i.e. `24h` \[default: 60m\] from a query and is invoked by GET request to
http://gateway-url:8080/function/system-metrics?function=myFunction&metrics_window=24h

> Note: If you're running the function on Swarm, you should update `prometheus_host` in `gateway_config.yml` to `prometheus`, i.e. removing the namespace suffix for Kubernetes.

> Note: `system-metrics` function requires prometheus service to be up and running. If you scale down prometheus to 0, this will result in function invocation count not being updated and an error invoking `system-metrics`: `Couldn't get metrics from Prometheus for function...`


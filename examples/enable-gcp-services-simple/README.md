# enable-gcp-services: Simple Example

### Overview

In this example, we will see how to generate [project services](https://cloud.google.com/config-connector/docs/reference/resource-docs/serviceusage/service) for `compute.googleapis.com` and `redis.googleapis.com`.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/enable-gcp-services-simple
```

### Function invocation

Render the configuration, `enable-gcp-services` has been defined declaratively in the Kptfile and configured via `services.yaml`.

```shell
$ kpt fn render enable-gcp-services-simple
```

### Expected result

```shell
[RUNNING] "gcr.io/kpt-fn/enable-gcp-services:unstable"
[PASS] "gcr.io/kpt-fn/enable-gcp-services:unstable"
  Results:
    [INFO] generated service in object "serviceusage.cnrm.cloud.google.com/v1beta1/Service/proj1-service-compute"
    [INFO] generated service in object "serviceusage.cnrm.cloud.google.com/v1beta1/Service/proj1-service-redis"

Successfully executed 1 function(s) in 1 package(s).
```

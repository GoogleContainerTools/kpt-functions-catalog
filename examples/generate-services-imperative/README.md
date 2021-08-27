# generate-services: Imperative Example

### Overview

This examples shows how to generate [Config Controller Service resources](https://cloud.google.com/config-connector/docs/reference/resource-docs/serviceusage/service) given a set Config Controller resources in the package by running [`generate-services`] function imperatively.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/generate-services-imperative
```

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn eval generate-services-imperative --image gcr.io/kpt-fn/generate-services-imperative:v0.1
```

### Expected result

There should be a sub-directory [`gcp-services`] created in the package with 4 yamls of `kind: Service`

[`generate-services`]: https://catalog.kpt.dev/generate-services/v0.1/

# set-namespace: Imperative Example

### Overview

This examples shows how to set namespace in the `.metadata.namespace` field on
all resources by running [`set-namespace`] function imperatively. Resources that
are known to be cluster-scoped will be skipped.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-namespace-imperative@set-namespace/v0.1.4
```

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn eval set-namespace-imperative --image gcr.io/kpt-fn/set-namespace:v0.1.4 -- namespace=example-ns
```

The desired namespace is provided after `--` and it will be converted to
`ConfigMap` by kpt and used as the function configuration.

### Expected result

Check all resources have `metadata.namespace` set to `example-ns`:

[`set-namespace`]: https://catalog.kpt.dev/set-namespace/v0.1/

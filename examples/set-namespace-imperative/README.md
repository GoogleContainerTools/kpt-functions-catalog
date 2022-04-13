# set-namespace: Imperative Example

### Overview

This examples shows how to replace KRM resources' namespace fields by matching
the namespace with the namespace object's `metadata.name`.  
The example uses `kpt fn eval` to run the function imperatively.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-namespace-imperative@set-namespace/v0.3.4
```

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn eval set-namespace-imperative --image gcr.io/kpt-fn/set-namespace:v0.3.4 -- namespace=example-ns
```

The desired namespace is provided after `--` and it will be converted to
`ConfigMap` by kpt and used as the function configuration.

### Expected result

Only the namespace fields which match the `namespace` "example" object has the value updated to
`new-ns`:

[`set-namespace`]: https://catalog.kpt.dev/set-namespace/v0.3/

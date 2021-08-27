# annotate-mutations: Simple Example

### Overview

This example shows how the [`annotate-mutations`] function works.

Running `annotate-mutations` function on the example packed will:

1.  Set `config.k8s.io/apply-time-mutation` annotation on resources with `apply-time-mutation` inline comments.

### Fetch the example package

Get the example package by running the following commands:

```shell
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/annotate-mutations-simple
```

### Function invocation

Invoke the function with the following command:

```shell
kpt fn eval annotate-mutations-simple --image gcr.io/kpt-fn/annotate-mutations:unstable
```

### Expected result

1.  File resources.yaml will include `config.k8s.io/apply-time-mutation` annotation matching the comment markups.

# annotate-apply-time-mutations: Simple Example

### Overview

This example shows how the [annotate-apply-time-mutations] function works.

Running `annotate-apply-time-mutations` function on the example package will:

1.  Set `config.k8s.io/apply-time-mutation` annotation on resources with `apply-time-mutation` inline comments.

### Fetch the example package

Get the example package by running the following commands:

```shell
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/contrib/examples/annotate-apply-time-mutations-simple
```

### Function invocation

Invoke the function with the following command:

```shell
kpt fn eval annotate-apply-time-mutations-simple --image gcr.io/kpt-fn-contrib/annotate-apply-time-mutations:unstable
```

### Expected result

1.  File resources.yaml will include `config.k8s.io/apply-time-mutation` annotation matching the comment markups.
2.  Commented fields with templated values will be updated with the template and replacement tokens.

[annotate-apply-time-mutations] https://catalog.kpt.dev/annotate-apply-time-mutations/v0.1/?id=definitions

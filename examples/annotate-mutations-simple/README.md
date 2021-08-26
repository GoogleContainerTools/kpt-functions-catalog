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
kpt fn eval annotate-mutations-simple --include-meta-resources --image gcr.io/kpt-fn/annotate-mutations:unstable
```

### Expected result

1.  File setters.yaml will include `project-id: foo` setter.
2.  In resources.yaml `my-test-project-second-bucket` StorageBucket resource
    will include `cnrm.cloud.google.com/project-id: foo` annotation.

[`annotate-mutations`]: https://catalog.kpt.dev/annotate-mutations/v0.1/

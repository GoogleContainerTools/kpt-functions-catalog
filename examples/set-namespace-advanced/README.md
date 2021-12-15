# set-namespace: Advanced Example

### Overview

This example demonstrates how to declaratively run [`set-namespace`] function to
add or replace the `.metadata.namespace` field on all resources except for those
known to be cluster-scoped.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-namespace-advanced@set-namespace/v0.1.4
```

We use the following `Kptfile` and `fn-config.yaml` to configure the function.

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/set-namespace:v0.1.4
      configPath: fn-config.yaml
```

```yaml
# fn-config.yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: SetNamespace
metadata:
  name: my-config
namespace: example-ns
additionalNamespaceFields:
  - group: dev.example.com
    version: v1
    kind: MyResource
    path: spec/configmapRef/namespace
    create: true
```

`set-namespace` function not only support `ConfigMap` but also a custom resource
of kind `SetNamespace` as the function configuration. In the `Kptfile`, we
specify `configPath` to point to the `functionConfig` file. The desired
namespace is provided using the `namespace` field in the function configuration.

If you have a resource that uses namespace in fields other
than `metadata.namespace`, you can use
`additionalNamespaceFields` to update those namespace fields.

In this example, we want the namespace to be set in
field `spec.configmapRef.namespace` in resources of kind `MyResource` as well.
We specify it in field `additionalNamespaceFields`.

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render set-namespace-advanced
```

### Expected result

Check all resources have `.metadata.namespace` set to `example-ns`:

[`set-namespace`]: https://catalog.kpt.dev/set-namespace/v0.1/

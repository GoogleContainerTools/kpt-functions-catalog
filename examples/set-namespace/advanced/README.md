# set-namespace: Advanced Example

### Overview

This example demonstrates how to declaratively run [`set-namespace`] function
to add or replace the `.metadata.namespace` field on all resources except for
those known to be cluster-scoped.

We use the following `Kptfile` and `fn-config.yaml` to configure the function.

```yaml
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: example
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/set-namespace:unstable
      configPath: fn-config.yaml
```

```yaml
# fn-config.yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: SetNamespaceConfig
metadata:
  name: my-config
namespace: example-ns
fieldSpecs:
  - group: dev.example.com
    version: v1
    kind: MyResource
    path: spec/configmapRef/namespace
    create: true
```

`set-namespace` function not only support `ConfigMap` but also a CRD as the
function configuration. We embed the CRD in the `Kptfile` in this example.
The desired namespace is provided using `.namespace` field in the function
configuration.

Suppose we have a CRD with group `dev.example.com`, version `v1` and kind
`MyResource`. We want the namespace to be set in field
`.spec.configmapRef.namespace` as well. We specify it in field `fieldSpecs`.

### Function invocation

Get the example config and try it out by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-namespace/advanced
$ kpt fn render advanced
```

### Expected result

Check all resources have `.metadata.namespace` set to `example-ns`:

[`set-namespace`]: https://catalog.kpt.dev/set-namespace/v0.1/

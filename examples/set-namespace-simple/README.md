# set-namespace: Simple Example

### Overview

This example demonstrates how to run [`set-namespace`] function
to replace the  `namespace` resource type in a variety of KRM resources.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-namespace-simple@set-namespace/v0.4.0
```

We use the following `Kptfile` to configure the function.

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/set-namespace:v0.4.0
      configMap:
        namespace: example-ns
```

The function configuration is provided using a `ConfigMap`. We set only one
key-value pair:
- `namespace: example-ns`: The desired namespace.

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render set-namespace-simple
```

### Expected result

Check all resources have `metadata.namespace` set to `example-ns`

[`set-namespace`]: https://catalog.kpt.dev/set-namespace/v0.4/

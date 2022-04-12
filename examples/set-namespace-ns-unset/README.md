# set-namespace: ns Unset Example

### Overview

This example demonstrates how to run [`set-namespace`] function
to replace the  `namespace` resource type in a variety of KRM resources.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-namespace-ns-unset
```

We use the following `Kptfile` to configure the function.

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/set-namespace:unstable
      configMap:
        namespace: example-ns
```

The function does not have input resources. So it cannot find previous set namespace. It will return 
the error in resourceList.results.

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render set-namespace-ns-unset
```

### Expected result

Users should receive error saying:
  "could not find any namespace fields to update. This function requires at least one of Namespace objects or "
  "namespace-scoped resources to have their namespace field set."

[`set-namespace`]: https://catalog.kpt.dev/set-namespace/v0.1/

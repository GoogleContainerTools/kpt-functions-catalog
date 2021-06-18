# set-labels: Simple Example

### Overview

This example demonstrates how to declaratively run [`set-labels`] function
to upsert labels to the `.metadata.labels` field on all resources.

We use the following `Kptfile` to configure the function.

```yaml
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: example
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/set-labels:unstable
      configMap:
        color: orange
        fruit: apple
```

The desired labels are provided as key-value pairs through `ConfigMap`.

### Function invocation

Get the example config and try it out by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-labels/simple
$ kpt fn render simple
```

### Expected result

Check all resources have 2 labels `color: orange` and `fruit: apple`.

[`set-labels`]: https://catalog.kpt.dev/set-labels/v0.1/

# set-annotations: Simple Example

### Overview

This example demonstrates how to declaratively run [`set-annotations`] function
to upsert annotations to the `.metadata.annotations` field on all resources.

We use the following `Kptfile` to configure the function.

```yaml
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: example
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/set-annotations:unstable
      configMap:
        color: orange
        fruit: apple
```

The desired annotations are provided as key-value pairs through `ConfigMap`.

### Function invocation

Get the example config and try it out by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-annotations/simple
$ kpt fn render simple
```

### Expected result

Check the 2 annotations have been added.

[`set-annotations`]: https://catalog.kpt.dev/set-annotations/v0.1/

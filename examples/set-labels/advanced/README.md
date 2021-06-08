# set-labels: Advanced Example

### Overview

This example demonstrates how to declaratively run [`set-labels`] function
to upsert labels to the `.metadata.labels` field on all resources.

We use the following `Kptfile` and `fn-config.yaml` to configure the function.

```yaml
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: example
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/set-labels:unstable
      configPath: fn-config.yaml
```

```yaml
# fn-config.yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: SetLabelConfig
metadata:
  name: my-config
labels:
  color: orange
  fruit: apple
fieldSpecs:
  - kind: MyResource
    group: dev.example.com
    version: v1
    create: true
    path: spec/selector/labels
```

The desired labels is provided using `labels` field. We have a CRD with group
`dev.example.com`, version `v1` and kind `MyResource`. We want the labels to be
added to field `.spec.selector.labels` as well. We specify it in field
`fieldSpecs`.

### Function invocation

Get the example config and try it out by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-labels/advanced
$ kpt fn render advanced
```

### Expected result

Check all resources have 2 labels: `color: orange` and `fruit: apple`.

[`set-labels`]: https://catalog.kpt.dev/set-labels/v0.1/

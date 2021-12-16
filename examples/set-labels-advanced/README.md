# set-labels: Advanced Example

### Overview

This example demonstrates how to declaratively run [`set-labels`] function
to upsert labels to the `.metadata.labels` field on all resources.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-labels-advanced@set-labels/v0.1.5
```

We use the following `Kptfile` and `fn-config.yaml` to configure the function.

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/set-labels:v0.1.5
      configPath: fn-config.yaml
```

```yaml
# fn-config.yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: SetLabels
metadata:
  name: my-config
labels:
  color: orange
  fruit: apple
additionalLabelFields:
  - kind: MyResource
    group: dev.example.com
    version: v1
    create: true
    path: spec/selector/labels
```

The desired labels is provided using `labels` field. We have a CRD with group
`dev.example.com`, version `v1` and kind `MyResource`. We want the labels to be
added to field `.spec.selector.labels` as well. We specify it in field
`additionalLabelFields`.

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render set-labels-advanced
```

### Expected result

Check all resources have 2 labels: `color: orange` and `fruit: apple`. And the
resource of kind `MyResource` also has these 2 labels in `spec.selector.labels`.

[`set-labels`]: https://catalog.kpt.dev/set-labels/v0.1/

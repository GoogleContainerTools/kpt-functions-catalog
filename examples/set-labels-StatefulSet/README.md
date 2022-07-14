# set-labels: StatefulSet Example

### Overview

This example demonstrates how to declaratively run [`set-labels`] function
to upsert labels to the `.metadata.labels` field on all resources.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-labels-StatefulSet
```

We use the following `Kptfile` and `fn-config.yaml` to configure the function.

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/set-labels:unstable
      configPath: fn-config.yaml
      selectors:
        - kind: StatefulSet
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
```

The desired labels is provided using `labels` field. We have a CRD with group
`apps` and kind `StatefulSet`. 

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render set-labels-StatefulSet
```

### Expected result

Check all resources have 2 labels: `color: orange` and `fruit: apple`. Especially for each resource in slice 
under path `spec.volumeClaimTemplates[].metatdata.label`

[`set-labels`]: https://catalog.kpt.dev/set-labels/v0.1/

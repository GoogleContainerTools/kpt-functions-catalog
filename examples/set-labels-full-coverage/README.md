# set-labels: Full Coverage Example

### Overview

This example demonstrates how to declaratively run [`set-labels`] function
to upsert all common labels to different type of resources, as well as self-defined resource.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-labels-full-coverage
```

We use the following `Kptfile` and `fn-config.yaml` to configure the function.

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
  annotations:
    config.kubernetes.io/local-config: "true"
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/set-labels:unstable
      configPath: fn-config.yaml
```

```yaml
# fn-config.yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: SetLabels
metadata:
  name: my-config
  annotations:
    config.kubernetes.io/local-config: "true"
labels:
  color: orange
  fruit: apple
  app: new
```

The desired labels is provided using `labels` field. We have a CRD with group
`dev.example.com`, version `v1` and kind `MyResource`. 

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render set-labels-full-coverage
```

### Expected result

Check all resources, if the labels to the path should be upsert, 
then we can see 3 labels have been updated, `color: orange`, `fruit: apple` and `app: new`. 
If the path does not need to create, just update labels if the key already exists, otherwise, ignore. 

[`set-labels`]: https://catalog.kpt.dev/set-labels/v0.1/

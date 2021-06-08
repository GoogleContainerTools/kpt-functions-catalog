# set-annotations: Advanced Example

### Overview

This example demonstrates how to declaratively run [`set-annotations`] function
to upsert annotations to the `.metadata.annotations` field on all resources.

We use the following `Kptfile` and `fn-config.yaml` to configure the function.

```yaml
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: example
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/set-annotations:unstable
      configPath: fn-config.yaml
```

```yaml
# fn-config.yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: SetAnnotationConfig
metadata:
  name: my-func-config
annotations:
  fruit: apple
  color: orange
fieldSpecs:
  - kind: MyResource
    group: dev.example.com
    version: v1
    create: true
    path: spec/selector/annotations
```

The desired annotations are provided using the `annotations` field. We have a
CRD with group `dev.example.com`, version `v1` and kind `MyResource`. We want
the annotations to be added to field `.spec.selector.annotations` as well. We
specify it in field `fieldSpecs`.

### Function invocation

Get the example config and try it out by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-annotations/advanced
$ kpt fn render advanced
```

### Expected result

Check the 2 annotations have been added to both the k8s built-in resources and
the custom resources.

[`set-annotations`]: https://catalog.kpt.dev/set-annotations/v0.1/

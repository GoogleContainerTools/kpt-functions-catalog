# set-annotation: Advanced Example

The `set-annotation` function adds annotations to KRM resources.

We use the following `SetAnnotationConfig` to configure the function.

```yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: SetAnnotationConfig
metadata:
  ...
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

## Function invocation

Get the example config and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-annotation/advanced .
kpt fn run advanced
```

## Expected result

Check the 2 annotations have been added to both the k8s built-in resources and
the custom resources.

```sh
kpt cfg cat advanced
```

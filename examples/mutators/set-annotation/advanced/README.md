# Set Annotation Advanced Example

The `set-annotation` function adds annotations to KRM resources.

In this example, we use `SetAnnotationConfig` to configure the function. The
desired annotations are provided using the `annotations` field.

We also specify `fieldSpecs` for our CRD with group as `dev.example.com`,
version as `v1` and kind as `MyResource`. The annotations will also be added to
field `.spec.selector.annotations`.

```yaml
apiVersion: kpt.dev/v1beta1
kind: SetAnnotationConfig
metadata:
  ...
annotations:
  fruit: apple
  configmanagement.gke.io/namespace-selector: sre-supported
fieldSpecs:
  - kind: MyResource
    group: dev.example.com
    version: v1
    create: true
    path: spec/selector/annotations
```

## Function invocation

Get the example config and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/mutators/set-annotation/advanced .
kpt fn run advanced
```

## Expected result

Check the 2 annotations have been added to both the k8s built-in resources and
the custom resources.

```sh
kpt cfg cat advanced
```

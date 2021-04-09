# set-label: Advanced Example

The `set-label` function adds or updates labels in the `.metadata.labels` field
and other fields that has the same meaning as a label on all resources. You can
find more details about these fields in the help text of the function.

We use the following `SetLabelConfig` to configure the function.

```yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: SetLabelConfig
metadata:
  ...
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

## Function invocation

Get the example config and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/mutators/set-label/advanced .
kpt fn run advanced
```

## Expected result

Check all resources have 2 labels: `color: orange` and `fruit: apple`.

```sh
kpt cfg cat advanced
```

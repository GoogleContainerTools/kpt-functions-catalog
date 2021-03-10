# Set Label Advanced Example

The `set-label` function adds or updates labels in the `.metadata.labels` field
and other fields that has the same meaning as a label on all resources. You can
find more details about these fields in the help text of the function.

In this example, we use SetLabelConfig to configure the function. The desired
labels is provided using `labels` field.

We also specify `fieldSpecs` for our CRD with group as `dev.example.com`,
version as `v1` and kind as `MyResource`. The labels will also be added to
`.spec.selector.labels`.

```yaml
apiVersion: kpt.dev/v1beta1
kind: SetLabelConfig
metadata:
  ...
fieldSpecs:
  - kind: MyResource
    group: dev.example.com
    version: v1
    create: true
    path: spec/selector/labels
labels:
  color: orange
  fruit: apple
```

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

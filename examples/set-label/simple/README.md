# set-label: Simple Example

The `set-label` function adds or updates labels in the `.metadata.labels` field
and other fields that has the same meaning as a label on all resources. You can
find more details about these fields in the help text of the function.

We use the following ConfigMap to configure the function.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  ...
data:
  color: orange
  fruit: apple
```

The desired labels are provided as key-value pairs using `data` field.

## Function invocation

Get the example config and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-label/simple .
kpt fn run simple
```

## Expected result

Check all resources have 2 labels `color: orange` and `fruit: apple`.

```sh
kpt cfg cat simple
```

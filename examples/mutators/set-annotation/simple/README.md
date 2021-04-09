# set-annotation: Simple Example

The `set-annotation` function adds annotations to KRM resources.

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

The desired annotations are provided as key-value pairs using `data` field.

## Function invocation

Get the example config and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/mutators/set-annotation/simple@go/set-annotation/v0.1.1 .
kpt fn run simple
```

## Expected result

Check the 2 annotations have been added.

```sh
kpt cfg cat simple
```

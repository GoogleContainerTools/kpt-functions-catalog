# Set Annotation Simple Example

The `set-annotation` function adds annotations to KRM resources.

In this example, we use ConfigMap to configure the function. The desired
annotations are provided as key-value pairs using `data` field.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  ...
data:
  configmanagement.gke.io/namespace-selector: sre-supported
  fruit: apple
```

## Function invocation

Get the example config and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/mutators/set-annotation/simple .
kpt fn run simple
```

## Expected result

Check the 2 annotations have been added.

```sh
kpt cfg cat simple
```

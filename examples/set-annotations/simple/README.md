# set-annotations: Simple Example

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

<!-- @getAndRunPkg @test -->
```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-annotations/simple@set-annotations/v0.1 .
kpt fn run simple
```

## Expected result

Check the 2 annotations have been added.

```sh
kpt cfg cat simple
```

# Set Namespace Simple Example

The `set-namespace` function adds or replaces the `.metadata.namespace` field on
all resources except for those known to be cluster-scoped.

In this example, we use ConfigMap to configure the function. The desired
namespace is provided using `.data.namespace` field.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  ...
data:
  namespace: example-ns
```

## Function invocation

Get the config example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/mutators/set-namespace/simple .
kpt fn run simple
```

## Expected result

Check all resources have `metadata.namespace` set to `example-ns`:

```sh
kpt cfg cat simple
```

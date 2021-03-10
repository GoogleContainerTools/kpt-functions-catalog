# Set Namespace Advanced Example

The `set-namespace` function adds or replaces the `.metadata.namespace` field on
all resources except for those known to be cluster-scoped.

In this example, we use `SetNamespaceConfig` to configure the function. The
desired namespace is provided using `.data.namespace` field.

We also specify `fieldSpecs` for our CRD with group as `dev.example.com`,
version as `v1` and kind as `MyResource`. The namespace will also be added to
`.spec.selector.namespace`.

```yaml
apiVersion: kpt.dev/v1beta1
kind: SetNamespaceConfig
metadata:
  ...
namespace: example-ns
fieldSpecs:
  - group: dev.example.com
    version: v1
    kind: MyResource
    path: spec/selector/namespace
    create: true
```

## Function invocation

Get the example config and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/mutators/set-namespace/advanced .
kpt fn run advanced
```

## Expected result

Check all resources have `.metadata.namespace` set to `example-ns`:

```sh
kpt cfg cat advanced
```

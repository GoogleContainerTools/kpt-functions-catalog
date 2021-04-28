# set-namespace: Advanced Example

The `set-namespace` function adds or replaces the `.metadata.namespace` field on
all resources except for those known to be cluster-scoped.

We use the following `SetNamespaceConfig` to configure the function.

```yaml
apiVersion: fn.kpt.dev/v1alpha1
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

The desired namespace is provided using `.data.namespace` field. We have a CRD
with group `dev.example.com`, version `v1` and kind `MyResource`. We want the
namespace to be set in field `.spec.selector.annotations` as well. We specify it
in field `fieldSpecs`.

## Function invocation

Get the example config and try it out by running the following commands:

<!-- @getAndRunPkg @test -->
```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-namespace/advanced .
kpt fn run advanced
```

## Expected result

Check all resources have `.metadata.namespace` set to `example-ns`:

```sh
kpt cfg cat advanced
```

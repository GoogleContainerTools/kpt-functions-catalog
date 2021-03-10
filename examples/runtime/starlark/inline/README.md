# Starlark Inline Example

In this example, we are going to demonstrate how to execute a starlark script
in starlark function against the input KRM resources to update or validate them.

The starlark script is provided inline in the function configuration like below:

```
apiVersion: fn.kpt.dev/v1beta1
kind: StarlarkFunction
metadata:
  ...
source: |
  # set the namespace on each resource
  def run(r, ns_value):
    for resource in r:
      # mutate the resource
      resource["metadata"]["namespace"] = ns_value
  # get the value to add
  ns_value = ctx.resource_list["functionConfig"]["data"]["foo"]
  run(ctx.resource_list["items"], ns_value)
data:
  foo: baz
```

This script updates namespace for all resources.

## Function invocation

Get the config example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/runtime/starlark/inline .
kpt fn run inline
```

## Expected result

Check the `.metadata.namespace` field has been set to `baz` for every resource.

```sh
kpt cfg cat inline
```

# starlark: Inline Example

## Overview

In this example, we are going to demonstrate how to use the starlark function
with an inline starlark script as function configuration.

We are going to use the following function configuration:

```
apiVersion: fn.kpt.dev/v1alpha1
kind: StarlarkRun
metadata:
  ...
source: |
  # set the namespace on each resource
  def run(r, ns_value):
    for resource in r:
      # mutate the resource
      resource["metadata"]["namespace"] = ns_value
  run(ctx.resource_list["items"], "prod")
```

The starlark script is embedded in the `source` field. This script read the
input KRM resources from `ctx.resource_list` and sets the `.metadata.namespace`
to `prod` for all resources.

## Function invocation

Get the config example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/runtime/starlark/inline .
kpt fn run inline
```

## Expected result

Check the `.metadata.namespace` field has been set to `prod` for every resource.

```sh
kpt cfg cat inline
```

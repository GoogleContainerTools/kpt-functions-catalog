# starlark: Simple Example

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
  # set the namespace on all resources
  def setnamespace(resources, namespace):
    for resource in resources:
      # mutate the resource
      resource["metadata"]["namespace"] = namespace
  setnamespace(ctx.resource_list["items"], "prod")
```

The starlark script is embedded in the `source` field. This script read the
input KRM resources from `ctx.resource_list` and sets the `.metadata.namespace`
to `prod` for all resources.

## Function invocation

Get the config example and try it out by running the following commands:

<!-- @getAndRunPkg @test -->
```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/starlark/simple@starlark/v0.1 .
kpt fn run simple
```

## Expected result

Check the `.metadata.namespace` field has been set to `prod` for every resource.

```sh
kpt cfg cat simple
```

# starlark: Set Namespace

### Overview

In this example, we are going to demonstrate how to declaratively run the
[`starlark`] function with an inline starlark script as function configuration
to set namespaces to KRM resources.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/starlark-set-namespace@starlark/v0.1.2
```

We are going to use the following `Kptfile` and `fn-config.yaml` to configure
the function:

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/starlark:v0.1.2
      configPath: fn-config.yaml
```

```yaml
# fn-config.yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: StarlarkRun
metadata:
  name: set-namespace-to-prod
  annotations:
source: |
  # set the namespace on all resources
  def setnamespace(resources, namespace):
    for resource in resources:
      # mutate the resource
      resource["metadata"]["namespace"] = namespace
  setnamespace(ctx.resource_list["items"], "prod")
```

The Starlark script is embedded in the `source` field. This script reads the
input KRM resources from `ctx.resource_list` and sets the `.metadata.namespace`
to `prod` for all resources.

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render starlark-set-namespace
```

### Expected result

Check the `.metadata.namespace` field has been set to `prod` for every resource.

[`starlark`]: https://catalog.kpt.dev/starlark/v0.1/

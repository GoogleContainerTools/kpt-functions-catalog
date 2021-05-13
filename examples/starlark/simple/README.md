# starlark: Simple Example

### Overview

In this example, we are going to demonstrate how to declaratively run the
[`starlark`] function with an inline starlark script as function configuration.

We are going to use the following Kptfile to run the function:

```
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: example
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/starlark:unstable
      config:
        apiVersion: fn.kpt.dev/v1alpha1
        kind: StarlarkRun
        metadata:
          name: set-namespace-to-prod
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

### Function invocation

Get the config example and try it out by running the following commands:


```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/starlark/simple .
kpt fn render simple
```

### Expected result

Check the `.metadata.namespace` field has been set to `prod` for every resource.

```sh
kpt pkg cat simple
```

[`starlark`]: https://catalog.kpt.dev/starlark/v0.1/

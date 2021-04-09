# starlark

### Overview

<!--mdtogo:Short-->

Run a Starlark script to update or validate resources.

<!--mdtogo-->

### Synopsis

<!--mdtogo:Long-->

You can specify your Starlark script inline under field source like this:

```yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: StarlarkRun
metadata:
  name: my-star-fn
source: |
  # set the namespace on each resource
  def run(resources, ns_value):
    for resource in resources:
    # mutate the resource
    resource["metadata"]["namespace"] = ns_value
  run(ctx.resource_list["items"], "prod")
```

<!--mdtogo-->

### Examples

<!-- TODO: update the following link to web page -->

<!--mdtogo:Examples-->

https://github.com/GoogleContainerTools/kpt-functions-catalog/tree/master/examples/starlark/

<!--mdtogo-->

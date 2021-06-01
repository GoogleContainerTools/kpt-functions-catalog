# starlark

### Overview

<!--mdtogo:Short-->

Run a Starlark script to mutate or validate resources.

<!--mdtogo-->

### FunctionConfig

<!--mdtogo:Long-->

[Starlark], formerly known as Skylark, is a dialect of Python. It is commonly
used as a configuration language. It is an untyped dynamic language with
high-level data types, first-class functions with lexical scope, and garbage
collection.

The starlark function is light-weight approach to write functions. You will have
the following advantages:

- No need to maintain container images.
- No need to do serialization and deserialization.
- Only need a Starlark script.

The starlark function currently uses `StarlarkRun` as the function config. It
looks like this:

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

`StarlarkRun` has the following field beside the standard KRM fields:
- `source`: (required) The source code of the Starlark script.

<!--mdtogo-->

[Starlark]: https://docs.bazel.build/versions/master/skylark/language.html
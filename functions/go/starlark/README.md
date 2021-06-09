# starlark

### Overview

<!--mdtogo:Short-->

The `starlark` function contains an interpreter for the [Starlark] language. It
can run a Starlark script to mutate or validate resources. It follows the
[executable configuration] pattern.

[Starlark] is a dialect of Python. It is commonly used as a configuration
language. It is an untyped dynamic language with high-level data types,
first-class functions with lexical scope, and garbage collection. You can find
the spec for the Starlark language [here][spec]. You can also find its API
reference [here][apiref].

<!--mdtogo-->

<!--mdtogo:Long-->

### FunctionConfig

The starlark function accepts a CRD of kind `StarlarkRun` as the
`functionConfig`. It looks like this:

```yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: StarlarkRun
metadata:
  name: my-star-fn
source: |
  # Starlark source lives here.
  # set the namespace on each resource
  def run(resources, ns_value):
    for resource in resources:
    # mutate the resource
    resource["metadata"]["namespace"] = ns_value
  run(ctx.resource_list["items"], "prod")
```

`StarlarkRun` has the following field besides the standard KRM fields:
- `source`: (required) A multi-line string that contains the source code of the
  Starlark script.

Here's what you can do in the Starlark script:

- Read resources from `ctx.resource_list`. The `ctx.resource_list` complies with
  the [KRM Functions Specification]. You can read the input resources from
  `ctx.resource_list[items]` and the `functionConfig` from
  `ctx.resource_list[functionConfig]`.
- Write resources to `ctx.resource_list[items]`. We do NOT support the `results`
  field, i.e. if you write some results in `ctx.resource_list["results"]`, they
  will be ignored.
- Return an error using [`fail`][fail].
- Write error message to stderr using [`print`][print]

In Starlark, a [for loop] is permitted only within a function definition. It
means if you want to iterate over `ctx.resource_list["items"]`, it has to be in
a function. You can refer the example `functionConfig` above.

### Debugging

<!-- TODO: fix https://github.com/GoogleContainerTools/kpt/issues/2200 -->

It is possible to debug the `starlark` functions using [`print`][print].

For example, you can add something like the following in your Starlark script:

```python
print(ctx.resource_list["items"][0]["metadata"]["name"])
```

Then you can run the function:

```shell
kpt fn render --results-dir=results
```

You will find your debugging output in `functionResultList items.stderr`.

<!--mdtogo-->

[Starlark]: https://docs.bazel.build/versions/master/skylark/language.html
[executable configuration]: https://kpt.dev/book/05-developing-functions/04-executable-configuration
[spec]: https://github.com/bazelbuild/starlark/blob/master/spec.md
[apiref]: https://docs.bazel.build/versions/master/skylark/lib/skylark-overview.html
[KRM Functions Specification]: https://kpt.dev/book/05-developing-functions/01-functions-specification
[for loop]: https://github.com/bazelbuild/starlark/blob/master/spec.md#for-loops
[fail]: https://docs.bazel.build/versions/master/skylark/lib/globals.html#fail
[print]: https://docs.bazel.build/versions/master/skylark/lib/globals.html#print

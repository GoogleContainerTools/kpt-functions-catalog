# starlark

### Overview

<!--mdtogo:Short-->

Run a Starlark script to mutate or validate resources.

<!--mdtogo-->

[Starlark] is a dialect of Python. It is commonly used as a configuration
language. It is an untyped dynamic language with high-level data types,
first-class functions with lexical scope, and garbage collection. You can find
the spec for the Starlark language [here][spec]. You can also find its API
reference [here][apiref].

<!--mdtogo:Long-->

### Function

[starlark] function contains an interpreter for the Starlark language. You can
provide a Starlark script in the `functionConfig` to mutate or validate your
resources. It is a light-weight way to write a KRM function. It is recommended
to only build simple function with it. Generally, if you Starlark script is
longer than 20 lines, you may want to consider build a function with our
[Golang SDK] or [TypeScript SDK].

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
  # You can mutate or validate the resources in ctx.resource_list["items"].
```

`StarlarkRun` has the following field besides the standard KRM fields:
- `source`: (required) A multi-line string that contains the source code of the
  Starlark script.

<!--mdtogo-->

[Starlark]: https://docs.bazel.build/versions/master/skylark/language.html
[spec]: https://github.com/bazelbuild/starlark/blob/master/spec.md
[apiref]: https://docs.bazel.build/versions/master/skylark/lib/skylark-overview.html
[Golang SDK]: https://kpt.dev/book/05-developing-functions/02-developing-in-Go
[TypeScript SDK]: https://kpt.dev/book/05-developing-functions/03-developing-in-Typescript

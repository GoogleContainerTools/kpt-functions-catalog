# format: Imperative Example

>? This example only works with 1.0.0-beta.1 or lower versions of kpt. Starting
>  from 1.0.0-beta.2, the order of the fields is preserved by kpt CLI.

This example demonstrates how to imperatively invoke the [format] function to
format KRM resources.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/format-imperative@format/v0.1.0
```

This example depicts the functionality of `format` function by formatting a
`Deployment` resource.

## Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn eval format-imperative --image gcr.io/kpt-fn/format:v0.1.0
```

## Expected result

The fields should be ordered as per OpenAPI schema definition of `Deployment`
resource. For e.g. `metadata.name` field is placed before `metadata.labels`
field. The keys in `metadata.labels` field are sorted alphabetically.


[format]: https://catalog.kpt.dev/format/v0.1/
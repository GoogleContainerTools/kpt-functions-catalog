# format: Imperative Example

This example demonstrates how to imperatively invoke the [format] function to
format KRM resources.

The `format` function formats the field ordering in YAML configuration files.
This example depicts the functionality of `format` function by formatting a
`Deployment` resource.

## Function invocation

Get the config example and try it out by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/format/imperative
$ kpt fn eval imperative --image=gcr.io/kpt-fn/format:unstable
```

## Expected result

The fields should be ordered as per OpenAPI schema definition of `Deployment`
resource. For e.g. `metadata.name` field is placed before `metadata.labels`
field. The keys in `metadata.labels` field are sorted alphabetically.

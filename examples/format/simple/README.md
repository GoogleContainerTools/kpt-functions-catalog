# format: Simple Example

The `format` function formats the field ordering in YAML configuration files. This example depicts the functionality of
`format` function by formatting a `Deployment` resource.

## Function invocation

Get the config example and try it out by running the following commands:

<!-- @getAndRunPkg @test -->

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/format/simple .
kpt fn run simple
```

## Expected result

The fields should be ordered as per openapi schema definition of `Deployment` resource. For e.g. `metadata.name` field
is placed before `metadata.labels` field. The keys in `metadata.labels` field are sorted alphabetically.

Verify that the changes are as described:

```sh
$ kpt cfg cat simple/
```

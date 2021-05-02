# format: Simple Example

The `format` function formats the field ordering in YAML configuration files.
Field ordering follows the ordering defined in the source Kubernetes resource definitions,
falling back on lexicographical sorting for unrecognized fields.

## Function invocation

Get the config example and try it out by running the following commands:

<!-- @getAndRunPkg @test -->

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/format/simple .
kpt fn run simple
```

## Expected result

Check the input resource is formatted as described:

```sh
$ kpt cfg cat simple/
```

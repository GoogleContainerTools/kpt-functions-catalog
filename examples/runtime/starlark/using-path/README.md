# Starlark Path Example

In this example, starlark source is provided using path in the function
configuration.

## Function invocation

Get the config example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/runtime/starlark/using-path .
kpt fn run --mount type=bind,src="$(pwd)/using-path/star-source",dst=/starlark/ using-path
```

## Expected result

Check the `.metadata.namespace` field has been set to `baz` for every resource.

```sh
kpt cfg cat using-path
```

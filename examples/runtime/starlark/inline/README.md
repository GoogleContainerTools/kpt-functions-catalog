# Starlark Inline Example

In this example, starlark source is provided inline in the function
configuration.

## Function invocation

Get the config example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/runtime/starlark/inline .
kpt fn run inline
```

## Expected result

Check the `.metadata.namespace` field has been set to `baz` for every resource.

```sh
kpt cfg cat inline
```

# Starlark URL Example

In this example, starlark source is provided using URL in the function
configuration.

Note: We need to set `network: true` in the `config.k8s.io/function` annotation.

## Function invocation

Get the config example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/runtime/starlark/using-url .
kpt fn run --network using-url
```

## Expected result

Check the `.metadata.namespace` field has been set to `baz` for every resource.

```sh
kpt cfg cat using-url
```

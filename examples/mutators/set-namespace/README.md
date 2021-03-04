# Set Namespace

The `set-namespace` KRM config function adds or replaces the
`metadata.namespace` field on all resources except for [those known to be
cluster-scoped].

## Function invocation

There 2 examples in this directory.

- simple: An example for simple function config format
- advanced: An example for advanced function config format

Get the simple config example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-namespace/simple .
kpt fn run simple --fn-path simple
```

## Expected result

Check all resources have `metadata.namespace` set to `example-ns`:

```sh
kpt cfg cat simple
```

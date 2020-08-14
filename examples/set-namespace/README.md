# Set Namespace

The `set-namespace` KRM config function adds or replaces the
`metadata.namespace` field on all resources. This example invokes the
set-namespace function using declarative configuration.

## Function invocation

Get this example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-namespace .
kpt fn run set-namespace/configs --fn-path set-namespace/functions
```

## Expected result

Check all resources have `metadata.namespace` set to `example-ns`:

```sh
kpt cfg cat set-namespace/configs
```

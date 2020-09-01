# Annotate Config

The `annotate-config` KRM config function adds an annotation to resources.

## Function invocation

Get this example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/annotate-config .
kpt fn run annotate-config/configs --fn-path annotate-config/functions
```

## Expected result

Check the `configmanagement.gke.io/namespace-selector` annotation was added:

```sh
kpt cfg cat annotate-config/configs
```

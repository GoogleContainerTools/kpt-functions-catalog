# Set Annotation

The `set-annotation` KRM config function adds an annotation to resources.

## Function invocation

There 2 examples in this directory.

- simple: An example for simple function config format
- advanced: An example for advanced function config format

Get the simple config example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-annotation/simple .
kpt fn run simple --fn-path simple/functions
```

## Expected result

Check the `configmanagement.gke.io/namespace-selector` annotation was added.
And annotations `fruit` is added to the custom resource:

```sh
kpt cfg cat simple
```

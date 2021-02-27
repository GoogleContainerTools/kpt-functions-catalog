# Set Label

The `set-label` KRM config function adds or replaces the
`metadata.labels` field on all resources except when users specify
their own selectors.

## Function invocation

There 2 examples in this directory.

- simple: An example for simple function config format
- advanced: An example for advanced function config format

Get the simple config example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-label/simple .
kpt fn run simple --fn-path simple/functions
```

## Expected result

Check all resources have a label `color` whose value is `orange`:

```sh
kpt cfg cat simple
```

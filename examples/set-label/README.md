# Set Label

The `set-label` KRM config function adds or replaces the
`metadata.labels` field on all resources except when users specify
their own selectors.

## Function invocation

There 2 examples in this directory.

- Simple function config format
- Complete function config format

Get the simple config example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-label/simple-config .
kpt fn run simple-config --fn-path simple-config/functions
```

## Expected result

Check all resources have a label `color` whose value is `orange`:

```sh
kpt cfg cat simple-config
```

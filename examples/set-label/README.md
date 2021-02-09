# Set Label

The `set-label` KRM config function adds or replaces the
`metadata.labels` field on all resources except when users specify
their own selectors.

## Function invocation

Get this example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-label .
kpt fn run set-label/configs --fn-path set-label/functions
```

## Expected result

Check all resources have a label `color` whose value is `orange`:

```sh
kpt cfg cat set-label/configs
```

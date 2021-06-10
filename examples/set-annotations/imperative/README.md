# set-annotations: Imperative Example

### Overview

This example demonstrates how to imperatively invoke [`set-annotations`]
function to upsert annotations to the `.metadata.annotations` field on all
resources.

### Function invocation

Get the example config and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-annotations/imperative
kpt fn eval imperative --image=gcr.io/kpt-fn/set-annotations:unstable -- color=orange fruit=apple
```

The labels provided in key-value pairs after `--` will be converted to a
`ConfigMap` by kpt and used as the function configuration.

### Expected result

Check the 2 annotations `color: orange` and `fruit: apple` have been added to
all resources.

[`set-annotations`]: https://catalog.kpt.dev/set-annotations/v0.1/

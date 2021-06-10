# set-labels: Imperative Example

### Overview

This example demonstrates how to declaratively run [`set-labels`] function to
upsert labels to the `.metadata.labels` field on all resources.

### Function invocation

Get the example config and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-labels/imperative
kpt fn eval imperative --image=gcr.io/kpt-fn/set-labels:unstable -- color=orange fruit=apple
```

The key-value pair(s) provided after `--` will be converted to `ConfigMap` by
kpt and used as the function configuration.

### Expected result

Check all resources have 2 labels `color: orange` and `fruit: apple`.

[`set-labels`]: https://catalog.kpt.dev/set-labels/v0.1/

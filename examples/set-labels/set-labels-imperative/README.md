# set-labels: Imperative Example

### Overview

This examples shows how to set labels in the `.metadata.labels` field on
all resources by running [`set-labels`] function imperatively.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-labels/set-labels-imperative
```

### Function invocation

Invoke the function by running the following commands:

```sh
$ kpt fn eval set-labels-imperative --image=gcr.io/kpt-fn/set-labels:unstable -- color=orange fruit=apple
```

The key-value pair(s) provided after `--` will be converted to `ConfigMap` by
kpt and used as the function configuration.

### Expected result

Check all resources have 2 labels `color: orange` and `fruit: apple`.

[`set-labels`]: https://catalog.kpt.dev/set-labels/v0.1/

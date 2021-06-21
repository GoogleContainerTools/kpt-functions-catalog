# set-annotations: Imperative Example

### Overview

This examples shows how to set annotations in the `.metadata.annotations` field
on all resources by running [`set-annotations`] function imperatively.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-annotations-imperative
```

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn eval set-annotations-imperative --image=gcr.io/kpt-fn/set-annotations:unstable -- color=orange fruit=apple
```

The labels provided in key-value pairs after `--` will be converted to a
`ConfigMap` by kpt and used as the function configuration.

### Expected result

Check the 2 annotations `color: orange` and `fruit: apple` have been added to
all resources.

[`set-annotations`]: https://catalog.kpt.dev/set-annotations/v0.1/

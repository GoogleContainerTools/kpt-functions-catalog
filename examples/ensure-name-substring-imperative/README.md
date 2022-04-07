# ensure-name-substring: Imperative Example

### Overview

This examples shows how to add prefix to resource names by
running [`ensure-name-substring`] function imperatively.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/ensure-name-substring-imperative@ensure-name-substring/v0.2.0
```

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn eval ensure-name-substring-imperative --image gcr.io/kpt-fn/ensure-name-substring:v0.2.0 -- prepend=prod-
```

The key-value pair(s) provided after `--` will be converted to `ConfigMap` by
kpt and used as the function configuration.

### Expected result

Check all resources have `prod-` prefix in their names:

[ensure-name-substring]: https://catalog.kpt.dev/ensure-name-substring/v0.2/

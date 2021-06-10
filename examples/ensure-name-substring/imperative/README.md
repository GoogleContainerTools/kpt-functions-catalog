# ensure-name-substring: Imperative Example

### Overview

This example demonstrates how to imperatively invoke the [`ensure-name-substring`]
function to prepend prefix in the resource names.

### Function invocation

Get the config example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/ensure-name-substring/imperative
kpt fn eval imperative --image=gcr.io/kpt-fn/ensure-name-substring:unstable -- prepend=prod
```

The key-value pair(s) provided after `--` will be converted to ConfigMap by kpt
and used as the function configuration.

### Expected result

Check all resources have `prod-` prefix in their names:

[issue]: https://github.com/GoogleContainerTools/kpt/issues/new/choose
[ensure-name-substring]: https://catalog.kpt.dev/ensure-name-substring/v0.1/

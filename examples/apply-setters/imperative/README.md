# apply-setters: Imperative Example

### Overview

<!-- TODO(phanimarupaka): Populate this and below -->

### Function invocation

Get the config example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/apply-setters/imperative .
kpt fn eval imperative --image=gcr.io/kpt-fn/apply-setters:unstable -- name=my-new-map
```

The key-value pair(s) provided after `--` will be converted to ConfigMap by kpt
and used as the function configuration.

### Expected result

Check the value of setter `name` is set to `my-new-map`.
Check the value of setter `env` is set to array value `[prod, stage]`.

```sh
$ kpt pkg cat imperative/
```

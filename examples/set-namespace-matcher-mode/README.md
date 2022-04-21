# set-namespace: Matcher Mode Example

### Overview

This example demonstrates how to run [`set-namespace`] function in Matcher Mode.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-namespace-matcher-mode
```

Take a look at the `set-namespace-matcher-mode/Kptfile`.

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: set-namespace-matcher-mode
...
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/set-namespace:unstable
      configMap:
        namespace: new-example
        namespaceMatcher: example1
```
The resources.yaml has two namespace values `example1` and `example2`. Only the `example1` should be updated.

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render set-namespace-matcher-mode
```

### Expected result

The output should be
```bash
Package "set-namespace-matcher-mode": 
[RUNNING] "gcr.io/kpt-fn/set-namespace:unstable"
[PASS] "gcr.io/kpt-fn/set-namespace:unstable" in 600ms
  Results:
    [info]: namespace "example1" updated to "new-example", 3 values changed

Successfully executed 1 function(s) in 1 package(s).
```
Check the resources.yaml, only `example1` is changed to `new-example`.

[`set-namespace`]: https://catalog.kpt.dev/set-namespace/v0.1/

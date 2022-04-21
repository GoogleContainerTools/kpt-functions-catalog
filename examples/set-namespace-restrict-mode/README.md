# set-namespace: Restrict Mode Example

### Overview

This example demonstrates how to run [`set-namespace`] function in Restrict Mode.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-namespace-restrict-mode
```

Take a look at the `set-namespace-restrict-mode/Kptfile`

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: set-namespace-restrict-mode
...
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/set-namespace:unstable
      configMap:
        namespace: example-ns
```

Take a look at the set-namespace-restrict-mode/resources.yaml, all namespace fields are `example`

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render set-namespace-restrict-mode
```

### Expected result

The output should be 
```bash
Package "set-namespace-restrict-mode": 
[RUNNING] "gcr.io/kpt-fn/set-namespace:unstable"
[PASS] "gcr.io/kpt-fn/set-namespace:unstable" in 500ms
  Results:
    [info]: namespace "example" updated to "example-ns", 4 values changed

Successfully executed 1 function(s) in 1 package(s).
```
Check all resources have `metadata.namespace` set to `example-ns`

[`set-namespace`]: https://catalog.kpt.dev/set-namespace/v0.1/

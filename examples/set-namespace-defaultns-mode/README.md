# set-namespace: Defaultns Mode Example

### Overview

This example demonstrates how to run [`set-namespace`] function [in defaultns mode]. Only the namespace field
which matches the Namespace object's `metadata.name` will be updated.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-namespace-defaultns-mode
```
Below is the `set-namespace-defaultns-mode/Kptfile` which sets the new namespace to `example-ns`
```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: set-namespace-defaultns-mode
...
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/set-namespace:unstable
      configMap:
        namespace: example-ns
```

Take a look at the `set-namespace-defaultns-mode/resources.yaml`. 
Two namespace values exist, one is `example` and the other is `irrelevant`. The Namespace object is `metadata.name`
is `example`.

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render set-namespace-defaultns-mode
```

### Expected result

The stdout is
```bash
Package "set-namespace-defaultns-mode": 
[RUNNING] "gcr.io/kpt-fn/set-namespace:unstable"
[PASS] "gcr.io/kpt-fn/set-namespace:unstable" in 5.3s
  Results:
    [info]: namespace "example" updated to "example-ns", 3 values changed

Successfully executed 1 function(s) in 1 package(s).
```

Check the `resources.yaml`, all the `example` namespaces are changed to `example-ns`, `irrelevant` namespace are unchanged.

[`set-namespace`]: https://catalog.kpt.dev/set-namespace/v0.3
[in defaultns mode]: https://catalog.kpt.dev/set-namespace/v0.3/?id=defaultnamespace-mode
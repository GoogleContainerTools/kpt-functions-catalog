# set-namespace: KPT Package context Example

### Overview

This example demonstrates how to run [`set-namespace`] function with kpt variant constructor.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get --for-deployment https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-namespace-kpt-package-context
```

Since we use flag --for-deployment, kpt generates a local file `package-context.yaml` as below
```
apiVersion: v1
kind: ConfigMap
metadata:
  name: kptfile.kpt.dev
  annotations:
    config.kubernetes.io/local-config: "true"
data:
  name: set-namespace-kpt-package-context
```
We can use this file as function config.

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/set-namespace:unstable
      configPath: package-context.yaml
```

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render set-namespace-kpt-package-context
```

### Expected result

Verify that all namespace is updated from "old" to "example".

[`set-namespace`]: https://catalog.kpt.dev/set-namespace/v0.1/

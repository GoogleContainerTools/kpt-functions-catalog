# list-setters: Simple Example

### Overview

In this example, we will see how to list setters in a package.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/list-setters-simple
```

### Function invocation

Invoke the function by running the following command:

```shell
$ kpt fn eval --image gcr.io/kpt-fn/list-setters:unstable --include-meta-resources
```

### Expected result

```shell
[RUNNING] "gcr.io/kpt-fn/list-setters:unstable"
[PASS] "gcr.io/kpt-fn/list-setters:unstable"
  Results:
    [INFO] Name: env, Value: [stage, dev], Type: list, Count: 1
    [INFO] Name: nginx-replicas, Value: 3, Type: string, Count: 1
    [INFO] Name: tag, Value: 1.16.2, Type: string, Count: 1
```

#### Note:

Refer to the `create-setters` function documentation for information about creating setters.

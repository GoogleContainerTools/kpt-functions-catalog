# put-setter-values: Simple Example

### Overview

In this example, we will see how to put desired setter values in the `setters.yaml`
file referenced from `Kptfile` via `configPath` option.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/put-setter-values-simple
```

### Function invocation

Invoke the function by running the following command:

```shell
$ kpt fn eval -i gcr.io/kpt-fn/put-setter-values:unstable --include-meta-resources \
-- namespace=my-space image=ubuntu env='[dev, prod]' tag=1.14.2
```

### Expected result

1. Check the value of field `data.image` is set to value `ubuntu` in `setters.yaml` file.
2. Check the new field `data.env` with value `[dev, prod]` is added to `setters.yaml` file.
3. Check the new field `data.tag` with value `1.14.2` is added to `setters.yaml` file.

### Next

This function will just declare the setter values. Please invoke `kpt fn render`
in order to render the resources with declared setter values.

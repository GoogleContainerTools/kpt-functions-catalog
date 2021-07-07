# put-setter-values

### Overview

<!--mdtogo:Short-->

Put setter values in Kptfile or functionConfig for setters.

<!--mdtogo-->

### FunctionConfig

<!--mdtogo:Long-->

This function is a porcelain to declare the setter values using imperative CLI command.
This function is meant to be used as an alternative to editing file directly to declare 
setter values before invoking [apply-setters] function.

Here is the recommended way to invoke this function

```shell
$ kpt fn eval -i gcr.io/kpt-fn/put-setter-values:unstable -- setter-name1=setter-value1 setter-name2=setter-value2
```

`put-setter-values` function performs the following steps when invoked:
1. Searches for `apply-setters` function declaration in the `Kptfile` present in root directory.
2. Updates the setter values either in the `configMap` section or in the file specified in `configPath` option.

<!--mdtogo-->

### Examples

<!--mdtogo:Examples-->

#### Setting scalar values

Let's start with the `Kptfile` and `setters.yaml` in a package

```yaml
# Kptfile
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: nginx
pipeline:
  mutators:
  - image: gcr.io/kpt-fn/apply-setters:v0.1
    configPath: setters.yaml
```

```yaml
# setters.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: setters
data:
  namespace: some-space
  image: nginx
  env: |
    - dev
    - stage
```

Invoke the function:

```shell
$ kpt fn eval -i gcr.io/kpt-fn/apply-setters:unstable -- image=ubuntu namespace=my-space env='[dev, prod]'
```

Modified setters.yaml looks like the following:

```yaml
# setters.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: setters
data:
  namespace: my-space
  image: ubuntu
  env: |
    - dev
    - prod
```

<!--mdtogo-->

### Next

This function will just declare the setter values. Please invoke `kpt fn render`
in order to render the resources with declared setter values.

[apply-setters]: https://catalog.kpt.dev/apply-setters/v0.1/

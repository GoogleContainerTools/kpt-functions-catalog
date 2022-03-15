# set-image: Simple Example

### Overview

This example demonstrates how to declaratively run [`set-image`] function
to set the `.spec.containers[].image` field to a specified tag on certain
resources.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-image-simple@set-image/v0.1.0
```

We use the following `Kptfile` to configure the function.

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
pipeline:
  mutators:
  - image: gcr.io/kpt-fn/set-image:v0.1.0
    configMap:
      name: nginx
      newName: bitnami/nginx
      newTag: 1.21.4
```

The desired image specification is provided through `ConfigMap` keys `name`,
`newName`, and `newTag`.

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render set-image-simple
```

### Expected result

Check the 2 images have been set to the specified tag.

[`set-image`]: https://catalog.kpt.dev/set-image/v0.1/

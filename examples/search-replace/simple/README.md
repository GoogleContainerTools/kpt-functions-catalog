# search-replace: Simple Example

### Overview

The `search-replace` function performs search and optionally replace fields
across all resources.

This is a simple example depicting search and replace operation on KRM resource config.

Let's start with the input resource

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: the-deployment
  namespace: my-space
```

Search matchers are provided with `by-` prefix. When multiple matchers are
provided they are AND’ed together. `put-` matchers are mutually exclusive.

We use the following `ConfigMap` to provide input matchers to the function.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
...
data:
  by-path: metadata.name
  by-value: the-deployment
  put-value: my-deployment
```

Invoking `search-replace` function would apply the changes to resource configs

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
  namespace: my-space
```

### Function invocation

Get the config example and try it out by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/search-replace/simple
$ kpt fn render simple
```

### Expected result

Check the value of deployment `the-deloyment` is changed to `my-deloyment`.

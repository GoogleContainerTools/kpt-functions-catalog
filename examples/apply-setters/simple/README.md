# apply-setters: Simple Example

### Overview

The `apply-setters` KRM config function applies setter values to resource fields
with setter references.

In this example, we will see how to apply desired setter values to the 
resource fields parameterized by `kpt-set` comments.

Let's start with the input resources

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: ubuntu-development # kpt-set: ${image}-development
data:
  some-key: some-value
---
apiVersion: v1
kind: MyKind
metadata:
  name: ubuntu # kpt-set: ${image}
image: nginx:1.1.2 # kpt-set: ${app}:${tag}
roles: # kpt-set: ${role}
  - dev
  - pro
```

We use `ConfigMap` to configure the `apply-setters` function. The desired
setter values are provided as key-value pairs using `data` field where key is
the name of the setter(as seen in the reference comments) and value is the new
desired value for the tagged field.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: apply-setters-fn-config
data:
  image: darwin
  role: |
    - dev
    - intermediate
  tag: 2.1.2
```

Invoking `apply-setters` function would apply the changes to resource configs

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: darwin-development # kpt-set: ${image}-development
data:
  some-key: some-value
---
apiVersion: v1
kind: MyKind
metadata:
  name: darwin # kpt-set: ${image}
image: nginx:2.1.2 # kpt-set: ${app}:${tag}
roles: # kpt-set: ${role}
  - dev
  - intermediate
```

### Function invocation

Get the config example and try it out by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/apply-setters/simple
$ kpt fn render simple
```

### Expected result

1. Check the value of field `metadata.name` is set to `darwin-development` in `ConfigMap` resource.
2. Check the value of field `metadata.name` is set to value `darwin` in `MyKind` resource.
3. Check the value of field `image` is set to value `nginx:2.1.2` in `MyKind` resource.
4. Check the value of field `roles` is set to array value `[dev, intermediate]` in `MyKind` resource.

#### Note:

Refer to the `create-setters` function documentation for information about creating setters.

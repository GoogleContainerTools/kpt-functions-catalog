# create-setters: Simple Example

### Overview

The `create-setters` KRM config function adds comments to resource fields
with setter references.

In this example, we will see how to add setter comments declaratively to
resource fields using `create-setters` function.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: ubuntu-development
data:
  some-key: some-value
---
apiVersion: v1
kind: MyKind
metadata:
  name: MyApp
image: nginx:1.1.2
roles:
  - dev
  - pro
```

We use `ConfigMap` to configure the `create-setters` function.
The desired setter values are provided as key-value pairs using `data` field.
Here, key is the name of the setter which is used to set the comment and value
is the field value to parameterize.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: create-setters-fn-config
data:
  role: |
    - dev
    - pro
  app: nginx
  image: ubuntu
  tag: 1.1.2
```

Invoking `create-setters` function would add the setter comments.

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
  name: ubuntu
image: nginx:1.1.2 # kpt-set: ${app}:${tag}
roles: # kpt-set: ${role}
  - dev
  - pro
```

#### Note:

If this function adds setter comments to fields for which you didn't intend to parameterize,
you can simply review and delete those comments manually.

### Function invocation

Get the config example and try it out by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/create-setters/simple
$ kpt fn render simple
```

### Expected result

Check the comment for resource with the value `ubuntu-development` is set to
`# kpt-set: ${image}-development` as it matches the setter `image: ubuntu`
Check the comment for resource with the value `nginx:1.1.2` is set to
`# kpt-set: ${app}:${tag}` as it matches the setters `image: ubuntu, tag: 1.1.2`
Check the comment for resource with the value `-dev\n-pro` is set to
`# kpt-set: ${role}` as it matches the array setter `role: |\n  - pro\n  - dev`
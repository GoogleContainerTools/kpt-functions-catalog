# create-setters: Simple Example

### Overview

The `create-setters` KRM config function adds comments to resource fields
with setter references.

In this example, we will see how to add setter comments declaratively to
resource fields using `create-setters` function.

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/create-setters-simple
```

```yaml
# resources.yaml
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
  name: ubuntu
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
# create-setters-fn-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: create-setters-fn-config
data:
  app: nginx
  image: ubuntu
  role: |
    - dev
    - pro
  tag: 1.1.2
```

Invoking `create-setters` function would add the setter comments.

```yaml
# resources.yaml
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

### Function invocation

Invoke the function by running the following commands:

```shell
$ kpt fn render create-setters-simple
```

### Expected result

`Comment` is added to the resources with the `Values` given below as they match the `Setters`.

| Setters                                    | Value                        | Comment                               |
|--------------------------------------------|------------------------------|---------------------------------------|
| <pre>image: ubuntu</pre>                   | <pre>ubuntu</pre>            | `# kpt-set: ${image}`                 |
| <pre>image: ubuntu</pre>                   | <pre>ubuntu-development</pre>| `# kpt-set: ${image}-development`     |
| <pre>app: nginx<br>tag: 1.1.2</pre>        | <pre>nginx:1.1.2</pre>       | `# kpt-set: ${app}:${tag}`            |
| <pre>role: \|<br>  - pro<br>  - dev</pre>  | <pre>- dev<br>- pro</pre>    | `# kpt-set: ${role}`                  |

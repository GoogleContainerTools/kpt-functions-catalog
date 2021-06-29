# create-setters: Simple Example

### Overview

In this example, we will see how to add [setter] comments to
resource fields using `create-setters` function.

### Fetch the example package

Get the example package by running the following command:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/create-setters-simple
```

We use `ConfigMap` to configure the `create-setters` function.
The desired setter values are provided as key-value pairs using `data` field.
Here, key is the name of the setter and value is the field value to be parameterized.

```yaml
# create-setters-fn-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: create-setters-fn-config
data:
  nginx-replicas: "4"
  env: |
    - dev
    - stage
  tag: 1.16.1
```

### Function invocation

Invoke the function by running the following command:

```shell
$ kpt fn render create-setters-simple
```

### Expected result

`Setter comment` is added to the resource with the `Value` given below as they match the `Setter`.

| Setter                                    | Value                        | Setter comment                               |
|--------------------------------------------|------------------------------|---------------------------------------|
| <pre>nginx-replicas: "4"</pre>  | <pre>4</pre>            | `# kpt-set: ${nginx-replicas}`                 |
| <pre>tag: 1.16.1</pre>        | <pre>nginx:1.16.1</pre>       | `# kpt-set: nginx:${tag}`            |
| <pre>env: <br>  - dev<br>  - stage</pre>  | <pre>- dev<br>- stage</pre>    | `# kpt-set: ${env}`                  |

[setter]: https://catalog.kpt.dev/apply-setters/v0.1/?id=setters-definition
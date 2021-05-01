# apply-setters: Simple Example

Setters provide a solution for template-free setting of field values. The
`apply-setters` KRM config function applies setter values to resource fields
with setter references.

Let's start with the input resources

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: the-map # kpt-set: ${name}
data:
  some-key: some-value
---
apiVersion: v1
kind: MyKind
metadata:
  name: ns
environments: # kpt-set: ${env}
  - dev
  - stage
```

We use ConfigMap to configure the `apply-setters` function. The desired
setter values are provided as key-value pairs using `data` field where key is
the name of the setter(as seen in the reference comments) and value is the new
desired value for the tagged field.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: apply-setters-fn-config
data:
  name: my-new-map
  env: |
    - prod
    - stage
```

Invoking `apply-setters` function would apply the changes to resource configs

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-new-map # kpt-set: ${name}
data:
  some-key: some-value
---
apiVersion: v1
kind: MyKind
metadata:
  name: ns
environments: # kpt-set: ${env}
  - prod
  - stage
```

## Function invocation

Get the config example and try it out by running the following commands:

<!-- @getAndRunPkg @test -->
```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/apply-setters/simple .
kpt fn run simple
```

## Expected result

Check the value of setter `name` is set to `my-new-map`.
Check the value of setter `env` is set to array value `[prod, stage]`.

```sh
$ kpt cfg cat simple/
```

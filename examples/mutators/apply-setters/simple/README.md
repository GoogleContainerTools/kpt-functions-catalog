# Apply Setters Example

The `apply-setters` KRM config function applies setter values to resource fields
with setter references.

In this example, we use ConfigMap to configure the function. The desired
setter values are provided as key-value pairs using `data` field.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  ...
data:
  name: my-new-map
  env: |
    - prod
    - stage
```

## Function invocation

Get the config example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/mutators/apply-setters/simple .
kpt fn run simple
```

## Expected result

Check the value of setter `name` is set to `my-new-map`.
Check the value of setter `env` is set to array value `[prod, stage]`.

```sh
kpt cfg cat simple
```

# ensure-name-substring: Advanced Example

In this example, we use the function `ensure-name-substring` to ensure every
resource name and the field declared in the field specs contain the desired name
substring. We prepend the substring if it doesn't exist.

We use the following CustomResource to configure the function.

```yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: EnsureNameSubstring
metadata:
  ...
substring: prod-
editMode: prepend
fieldSpecs:
  - group: dev.example.com
    version: v1
    kind: MyResource
    path: spec/name
```

The function will not only update field `.metadata.name` but also field
`.spec.name` in `MyResource`.

## Function invocation

Get the config example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/ensure-name-substring/advanced .
kpt fn run advanced
```

## Expected result

Check all resources have `prod-` in their names and the field `.spec.name` in
`MyResource` also got updated.

```sh
kpt cfg cat advanced
```

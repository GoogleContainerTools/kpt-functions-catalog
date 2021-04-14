# ensure-name-substring: Prefix Example

In this example, we use the function `ensure-name-substring` to ensure every
resource name contains the desired name substring. We prepend the substring if
it doesn't exist.

We use the following ConfigMap to configure the function.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  ...
data:
  prepend: prod-
```

## Function invocation

Get the config example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/ensure-name-substring/prefix .
kpt fn run prefix
```

## Expected result

Check all resources have `prod-` in their names:

```sh
kpt cfg cat prefix
```

We have a `Service` object whose name is `with-prod-service` which already
contains `prod-`. This function will skip it.

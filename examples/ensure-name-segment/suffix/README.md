# ensure-name-substring: Suffix Example

In this example, we use the function `ensure-name-substring` to ensure every
resource name contains the desired name substring. We append the substring if it
doesn't exist.

We use the following ConfigMap to configure the function.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  ...
data:
  append: -prod
```

## Function invocation

Get the config example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/ensure-name-substring/suffix .
kpt fn run suffix
```

## Expected result

Check all resources have `-prod` in their names:

```sh
kpt cfg cat suffix
```

We have a `Service` object whose name is `the-service-prod` which already
contains `-prod`. This function will skip it.

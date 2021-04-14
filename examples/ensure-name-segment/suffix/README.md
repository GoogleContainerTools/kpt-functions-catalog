# ensure-name-segment: Suffix Example

In this example, we use function `ensure-name-segment` to ensure every resource
name contains the desired name segment. We append the segment if it doesn't
exist.

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
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/ensure-name-segment/suffix .
kpt fn run suffix
```

## Expected result

Check all resources have `-prod` in their names:

```sh
kpt cfg cat suffix
```

We have a `Service` object whose name is `the-service-prod` which already
contains `-prod`. This function will skip it.

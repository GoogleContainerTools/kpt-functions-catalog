# ensure-name-substring: Suffix Example

Note: This is an alpha function, and we are actively seeking feedback on the
function config syntax and behavior. If you have suggestion or feedback, please
file an issue [here](https://github.com/GoogleContainerTools/kpt/issues/new/choose).

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

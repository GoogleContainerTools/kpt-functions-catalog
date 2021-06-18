# ensure-name-substring: Suffix Example

### Overview

This example demonstrates how to declaratively run the [`ensure-name-substring`]
function to append suffix in the resource names.

We use the following `Kptfile` to run the function.

```yaml
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: example
pipeline:
  mutators:
    - image: gcr.io/kpt-fn/ensure-name-substring:unstable
      configMap:
        append: -prod
```

We are going to append suffix `-prod` to resource names.

### Function invocation

Get the config example and try it out by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/ensure-name-substring/suffix
$ kpt fn render suffix
```

### Expected result

Check all resources have `-prod` in their names:

We have a `Service` object whose name is `the-service-prod` which already
contains substring`-prod`. This resource will be skipped.

[ensure-name-substring]: https://catalog.kpt.dev/ensure-name-substring/v0.1/

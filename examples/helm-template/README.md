# Helm Template

The `helm-template` KRM config function generates a new kpt package from a
local Helm chart. This example invokes the helm template function using
declarative configuration.

## Function invocation

Get this example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/helm-template .
kpt fn run helm-template/local-configs --mount type=bind,src=$(pwd)/helm-template/helloworld-chart,dst=/source
```

## Expected result

Verify the expanded configuration:

```sh
kpt cfg cat local-configs
```

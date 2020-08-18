# Kustomize Build

The `kustomize-build` config function generates a new kpt package per the
contents of a kustomization file. This example invokes the kustomize build
function using declarative configuration.

## Function invocation

Get this example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/kustomize-build .
kpt fn run kustomize-build/local-configs --mount type=bind,src="$(pwd)"/kustomize-build/kustomize-dir,dst=/source
```

## Expected result

Verify the expanded configuration:

```sh
kpt cfg cat kustomize-build/local-configs
```

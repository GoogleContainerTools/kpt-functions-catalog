# Kustomize Build

The `kustomize-build` config function generates a new kpt package per the contents of a
kustomization file. This example invokes the kustomize build function using declarative
configuration.

## Function invocation

The function is invoked using the function configuration in `local-configs/example.yaml`.

Get this example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/kustomize-build .
kpt fn run kustomize-build/local-configs --mount type=bind,src="$(pwd)"/kustomize-build/kustomize-dir,dst=/source
```

The first command fetches this example. The last command:

* loads kustomization file and resources from `kustomize-dir` into `/source` (from the `--mount` flag)
* reads configs from the `local-configs` folder (from invoking `kpt fn run` on `local-configs`)
* runs the function config from `local-configs/example.yaml` (from `example.yaml` containing the
  function annotation `config.kubernetes.io/function`)
* kustomizes the configuration from `/source` (from invoking the `gcr.io/kpt-functions/kustomize-build`
  container image)
* writes them back into the `local-configs` folder (also from invoking `kpt fn run` on
  `local-configs`)

Verify the expanded configuration:

```sh
kpt cfg cat kustomize-build/local-configs
```

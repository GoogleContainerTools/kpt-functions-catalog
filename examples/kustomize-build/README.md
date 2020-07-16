# Kustomize Build

This is an example of invoking the kustomize build function using declarative configuration.

## Function invocation

The function is invoked by authoring a function configuration in local-configs/example.yaml
with `metadata.annotations.[config.kubernetes.io/function]` set to
`gcr.io/kpt-functions/kustomize-build`.

Get this example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/kustomize-build .
cd kustomize-build; kpt cfg cat local-configs
kpt fn run local-configs --mount type=bind,src=$(pwd)/kustomize-dir,dst=/source
```

The first commands fetch this example and show that `local-configs` only contains `example.yaml`.

The last command:

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
kpt cfg cat local-configs
```

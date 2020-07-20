# Helm Template

The `helm-template` config function generates a new kpt package from a local Helm chart. This
example invokes the helm template function using declarative configuration.

## Function invocation

The function is invoked using the function configuration in `local-configs/example.yaml`.

Get this example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/helm-template .
kpt fn run helm-template/local-configs --mount type=bind,src=$(pwd)/helloworld-chart,dst=/source
```

The first commands fetches this example. The last command:

* loads the chart directory from `helloworld-chart` into `/source` (from the `--mount` flag)
* reads configs from the `local-configs` folder (from invoking `kpt fn run` on `local-configs`)
* runs the function config from `local-configs/example.yaml` (from `example.yaml` containing the
    function annotation `config.kubernetes.io/function`)
* expands the templates from `/source` (from invoking the `gcr.io/kpt-functions/helm-template`
    container image)
* writes them back into the `local-configs` folder (also from invoking `kpt fn run` on
    `local-configs`)

Verify the expanded configuration:

```sh
kpt cfg cat local-configs
```

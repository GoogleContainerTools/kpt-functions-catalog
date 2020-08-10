# Annotate Config

The `annotate-config` function adds an annotation to configuration. This
example invokes the annotate-config function using declarative configuration.

## Function invocation

The function is invoked using the function configuration in
`functions/fn-config.yaml`.

Get this example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/annotate-config .
kpt fn run annotate-config/configs --fn-path annotate-config/functions
```

The first command fetches this example. The last command:

* reads configs from the `annotate-config/configs` folder (from invoking
  `kpt fn run` on `annotate-config/configs`)
* runs the function config from `annotate-config/functions/fn-config.yaml`
  (since it contains the `config.kubernetes.io/function` function annotation)
* adds the given annotation to the configs from `annotate-config/configs`
* writes configs back into the `annotate-config/configs` folder

Check the `configmanagement.gke.io/namespace-selector` annotation was added:

```sh
kpt cfg cat annotate-config/configs
```

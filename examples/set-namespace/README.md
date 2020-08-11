# Set Namespace

The `set-namespace` config function sets the `metadata.namespace` field on all
resources. This example invokes the set-namespace function using declarative
configuration.

## Function invocation

The function is invoked using the function configuration in
`functions/fn-config.yaml`.

Get this example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/set-namespace .
kpt fn run set-namespace/configs --fn-path set-namespace/functions
```

The first command fetches this example. The last command:

* reads configs from the `set-namespace/configs` folder (from invoking
  `kpt fn run` on `set-namespace/configs`)
* runs the function config from `set-namespace/functions/fn-config.yaml`
  (since it contains the `config.kubernetes.io/function` function annotation)
* sets the namespace of the configs from `set-namespace/configs`
* writes configs back into the `set-namespace/configs` folder

Check the `example-ns` namespace was added:

```sh
kpt cfg cat set-namespace/configs
```

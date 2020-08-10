# Istioctl Analyze

The `istioctl-analyze` config function detects potential issues with your Istio
configuration and output errors as results. This example invokes the
istioctl-analyze function using declarative configuration.

## Function invocation

The function is invoked using the function configuration in
`functions/fn-config.yaml`.

Get this example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/istioctl-analyze .
kpt fn run istioctl-analyze/configs --fn-path istioctl-analyze/functions --results-dir istioctl-analyze/results
```

The first command fetches this example. The last command:

* reads configs from the `istioctl-analyze/configs` folder (from invoking
  `kpt fn run` on `istioctl-analyze/configs`)
* runs the function config from `istioctl-analyze/functions/fn-config.yaml`
  (since it contains the function annotation `config.kubernetes.io/function`)
* analyzes the configs from `istioctl-analyze/configs`
* writes configs back into the `istioctl-analyze/configs` folder

Check the results:

```sh
kpt cfg cat istioctl-analyze/results
```

The command results in an error
`Port name  (port: 5000, targetPort: 0) doesn't follow the naming convention of Istio port`.
Fix the error and rerun the `kpt fn run` command. This will return success (no
output).

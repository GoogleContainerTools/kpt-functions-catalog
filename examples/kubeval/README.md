# Kubeval

The `kubeval` KRM config function validates Kubernetes resources against their
Kubernetes OpenAPI definition using kubeval. This example invokes the kubeval
function against Kubernetes v1.18.0 using declarative configuration.

## Function invocation

Get this example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/kubeval .
kpt fn run kubeval --network
```

## Expected Results

The function outputs the following error:

```sh
[ERROR] Invalid type. Expected: [integer,null], given: string in object 'v1/ReplicationController//bob' in file configs/example-config.yaml
```

In the `configs/example-config.yaml` file, replace the value of `spec.replicas`
with an integer to pass validation and rerun the command. This will return
success (no output).

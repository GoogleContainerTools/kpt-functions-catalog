# Kubeval

The `kubeval` KRM config function validates Kubernetes resources using kubeval.
Learn more on the [kubeval website].

This example invokes the kubeval function against Kubernetes v1.18.0.

## Function invocation

Get this example and try it out by running the following commands:

<!-- TODO: https://github.com/GoogleContainerTools/kpt/issues/983 -->

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/kubeval .
kpt fn run kubeval --network
```

## Expected Results

This should give the following output:

```sh
[ERROR] Invalid type. Expected: [integer,null], given: string in object 'v1/ReplicationController//bob' in file configs/example-config.yaml in field spec.replicas
error: exit status 1
```

In the `configs/example-config.yaml` file, replace the value of `spec.replicas`
with an integer to pass validation and rerun the command. This will return
success (no output).

[kubeval website]: https://www.kubeval.com/

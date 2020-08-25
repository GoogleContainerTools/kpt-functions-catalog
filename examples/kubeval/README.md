# Kubeval

The `kubeval` KRM config function validates Kubernetes resources using kubeval.
Learn more on the [kubeval website].

This example invokes the kubeval function against Kubernetes v1.18.0 using
declarative configuration.

## Function invocation

Get this example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/kubeval .
kpt fn run kubeval --network --results-dir /tmp
```

## Expected Results

The `--results-dir` flag let us specify a destination to write function results
to. Check the results:

```sh
cat /tmp/results-0.yaml
```

They contain the following results:

```sh
- message: 'Invalid type. Expected: [integer,null], given: string'
  severity: error
  resourceRef:
    apiVersion: v1
    kind: ReplicationController
    namespace: ''
    name: bob
  file:
    path: configs/example-config.yaml
  field:
    path: spec.replicas
```

In the `configs/example-config.yaml` file, replace the value of `spec.replicas`
with an integer to pass validation and rerun the command. This will return
success (no output).

[kubeval website]: https://www.kubeval.com/

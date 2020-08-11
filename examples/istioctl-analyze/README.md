# Istioctl Analyze

The `istioctl-analyze` KRM config function detects potential issues with your
Istio configuration. It takes in a List of KRM configs sourced from a local
package and outputs validation errors by adding a `results` field to the List.
Kpt provides the `--results-dir` flag for users to specify a destination to
write these results to. This example invokes the istioctl-analyze function
using declarative configuration.

## Function Invocation

Get this example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/istioctl-analyze .
kpt fn run istioctl-analyze --results-dir /tmp
```

## Expected Results

Check the results:

```sh
kpt cfg cat /tmp/results-0.yaml
```

They contain the following error:

```sh
Port name  (port: 5000, targetPort: 0) doesn\'t follow the naming convention of Istio port
```

The error comes from the `v1/Service//helloworld` resource. Fix the error and
rerun the command. This will return success (no output).

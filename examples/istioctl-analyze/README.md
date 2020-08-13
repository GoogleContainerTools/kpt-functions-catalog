# Istioctl Analyze

The `istioctl-analyze` KRM config function detects potential issues with your
Istio configuration and outputs structured results detailing any errors found
during analysis. This example invokes the istioctl-analyze function using
declarative configuration.

## Function Invocation

Get this example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/istioctl-analyze .
kpt fn run istioctl-analyze --results-dir /tmp
```

## Expected Results

The `--results-dir` flag let us specify a destination to write function results
to. Check the results:

```sh
kpt cfg cat /tmp/results-0.yaml
```

They contain the following error:

```sh
Schema validation error: gateway must have at least one server
```

The error comes from the httpbin-gateway resource in
`configs/example-config.yaml`. Fix the error and rerun the command. This will
return success (no output).

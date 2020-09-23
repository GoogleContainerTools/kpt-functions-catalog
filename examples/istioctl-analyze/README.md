# Istioctl Analyze

The `istioctl-analyze` KRM config function detects potential issues with your
Istio configuration and outputs structured results detailing any errors found
during analysis. This example invokes the istioctl-analyze function using
declarative configuration.

## Function Invocation

Get this example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/istioctl-analyze .
kpt fn run istioctl-analyze
```

## Expected Results

This should give the following output:

```sh
[ERROR] Schema validation error: gateway must have at least one server in object 'networking.istio.io/v1alpha3/Gateway//httpbin-gateway' in file configs/example-config.yaml
error: exit status 1
```

The error comes from the httpbin-gateway resource in
`configs/example-config.yaml`. Uncomment `spec.servers` in that file to fix the
error and rerun the command. This will return success (no output).

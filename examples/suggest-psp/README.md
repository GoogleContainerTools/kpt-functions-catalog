# Suggest Pod Security Policy

The `suggest-psp` KRM config function lints pod security policies by suggesting
the 'spec.allowPrivilegeEscalation' field be set to 'false'. It outputs
structured results detailing which PSP objects should be changed. This example
invokes the suggest-psp function using declarative configuration.

## Function Invocation

Get this example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/suggest-psp .
kpt fn run suggest-psp --results-dir /tmp
```

## Expected Results

The `--results-dir` flag let us specify a destination to write function results
to. Check the results:

```sh
cat /tmp/results-0.yaml
```

They contain the following message:

```sh
Suggest explicitly disabling privilege escalation
```

The error comes from the psp resource in `configs/example-config.yaml`.
Uncomment `spec.allowPrivilegeEscalation` in that file to fix the
error and rerun the command. This will return success (no output).

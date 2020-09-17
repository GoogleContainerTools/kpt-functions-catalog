# Suggest Changes to `PodSecurityPolicy`

The `suggest-psp` KRM config function lints `PodSecurityPolicy` resources by
suggesting the 'spec.allowPrivilegeEscalation' field be set to 'false'. It
outputs structured results detailing which `PodSecurityPolicy` objects should be
changed.

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

They contain the following results:

```sh
- message: Suggest explicitly disabling privilege escalation
  severity: warn
  tags:
    category: security
  resourceRef:
    apiVersion: policy/v1beta1
    kind: PodSecurityPolicy
    namespace: ''
    name: psp
  file:
    path: configs/example-config.yaml
  field:
    path: spec.allowPrivilegeEscalation
    suggestedValue: false
```

The error comes from the psp resource in `configs/example-config.yaml`.
Uncomment `spec.allowPrivilegeEscalation` in that file to fix the error and
rerun the command. This will return success (no output).

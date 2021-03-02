# Suggest Changes to `PodSecurityPolicy`

The `suggest-psp` KRM config function lints `PodSecurityPolicy` resources by
suggesting the 'spec.allowPrivilegeEscalation' field be set to 'false'. It
outputs structured results detailing which `PodSecurityPolicy` objects should be
changed.

## Function Invocation

Get this example and try it out by running the following commands:

```sh
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/suggest-psp .
kpt fn run suggest-psp
```

## Expected Results

This should give the following output:

```sh
[WARN] Suggest explicitly disabling privilege escalation in object 'policy/v1beta1/PodSecurityPolicy//psp' in file configs/example-config.yaml
```

The error comes from the `psp` object in `configs/example-config.yaml`.
Uncomment `spec.allowPrivilegeEscalation` in that file to fix the error and
rerun the command. This will return success (no output).

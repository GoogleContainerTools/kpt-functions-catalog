# enforce-gatekeeper: warning only

## Overview

This example is very similar to the invalid configmap example. The major
difference is that the violations are warnings instead of errors.

In the constraint, we use `enforcementAction: warn` instead of
`enforcementAction: deny`.

## Function invocation

Get the package:

<!-- @getPkg @test -->
```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/enforce-gatekeeper/warning-only@enforce-gatekeeper/v0.1 .
```

Create a directory for storing the structured output.

```shell
$ cd warnning-only
$ mkdir results
```

Run the function:

```shell
$ kpt fn run --results-dir=results .
```

## Expected result

You won't any failure. But if you look at the structured output, you can find a
warning about the constraint violation.

```shell
$ cat results/results-0.yaml 
items:
- message: |-
    The following banned keys are being used in the ConfigMap map: {"private_key"}
    violatedConstraint: no-secrets-in-configmap
  severity: warning
  resourceRef:
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: super-secret
      namespace: default
  file:
    path: resources.yaml
    index: 2
```

You can find:
- a detailed error message
- what resource violates the constraints
- what constraint does it violate
- where does the resource live and its index in the file

To pass validation, let's replace the key `private_key` in the ConfigMap in
`resources.yaml` with something else e.g. `public_key`.
Rerun the command. It will no longer have the warning.

## Function Reference Doc

TODO: replace the following with the link to the reference doc when our site is live.
https://github.com/GoogleContainerTools/kpt-functions-catalog/blob/enforce-gatekeeper/v0.1/functions/go/enforce-gatekeeper/README.md

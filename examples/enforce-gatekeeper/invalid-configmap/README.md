# enforce-gatekeeper: invalid configmap

## Overview

This example demonstrates how to validate ConfigMaps using a constraint.

There are 3 resources: a ConstraintTemplate, a K8sBannedConfigMapKeysV1 and a
ConfigMap.
The constraint disallows using `private_key` as a key in the ConfigMap.

## Function invocation

Get the package:

<!-- @getPkg @test -->
```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/enforce-gatekeeper/invalid-configmap .
```

Create a directory for storing the structured output.

```shell
$ cd invalid-configmap
$ mkdir results
```

Run the function:

```shell
$ kpt fn run --results-dir=results .
```

## Expected result

You should see the following output:

```
The following banned keys are being used in the ConfigMap: {"private_key"}
violatedConstraint: no-secrets-in-configmaperror: exit status 1
```

Let's take a look at the structured output:

```shell
$ cat results/results-0.yaml 
items:
- message: |-
    The following banned keys are being used in the ConfigMap: {"private_key"}
    violatedConstraint: no-secrets-in-configmap
  severity: error
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
Rerun the command. It will succeed (no output).

## Function Reference Doc

TODO: replace the following with the link to the reference doc when our site is live.
https://github.com/GoogleContainerTools/kpt-functions-catalog/blob/master/functions/go/enforce-gatekeeper/README.md

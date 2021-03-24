# policy-controller-validate: invalid configmap

## Overview

This example demonstrates how to validate config maps using a constraint.

There are 3 resources: a ConstraintTemplate, a K8sBannedConfigMapKeysV1 and a
ConfigMap.
The constraint disallows `private_key` to be used as a key in the config map.

## Function invocation

```shell
kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/validators/policy-controller-validate/invalid-configmap .
kpt fn run invalid-configmap
```

## Expected result

It should complain like the following:

```
Found 1 violations:

[1] The following banned keys are being used in the config map: {"private_key"}

name: "super-secret"
path: resources.yaml
violatedConstraint: no-secrets-in-configmap

error: exit status 1
```

In the `resources.yaml` file, replace the key `private_key` in the config map
with something else e.g. `public_key` to pass validation.
Rerun the command. It will succeed (no output).

## Function Reference

TODO: add the link

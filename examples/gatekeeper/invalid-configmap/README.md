# gatekeeper: Invalid ConfigMap

### Overview

This example demonstrates how to declaratively run the [gatekeeper]
function to validate resources using gatekeeper constraints.

There are 3 resources: a `ConstraintTemplate`, a `K8sBannedConfigMapKeysV1` and
a `ConfigMap`.
The constraint disallows using `private_key` as a key in the `ConfigMap`.

Here's an example Kptfile to run the function:
```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
pipeline:
  validators:
    - image: gcr.io/kpt-fn/gatekeeper:unstable
```

### Function invocation

Get the package:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/gatekeeper/invalid-configmap
```

Run the function:

```shell
$ kpt fn render invalid-configmap --results-dir=/tmp
```

### Expected result

Let's take a look at the structured results in `/tmp/results.yaml`:

```yaml
apiVersion: kpt.dev/v1
kind: FunctionResultList
metadata:
  name: fnresults
exitCode: 1
items:
  - image: gcr.io/kpt-fn/gatekeeper:unstable
    stderr: |-
      The following banned keys are being used in the ConfigMap: {"private_key"}
      violatedConstraint: no-secrets-in-configmap
    exitCode: 1
    results:
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

To pass validation, let's replace the key `private_key` in the `ConfigMap` in
`resources.yaml` with something else e.g. `public_key`.
Rerun the command. It will succeed.

[gatekeeper]: https://catalog.kpt.dev/gatekeeper/v0.1/

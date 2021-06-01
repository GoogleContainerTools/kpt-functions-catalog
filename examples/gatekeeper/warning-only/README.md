# gatekeeper: Warning Only

### Overview

This example demonstrates how to declaratively run the [gatekeeper]
function to validate resources using gatekeeper constraints. The violations are
configured to be warnings instead of errors.

Here's an example `Kptfile` to run the function:

```yaml
apiVersion: kpt.dev/v1alpha2
kind: Kptfile
metadata:
  name: example
pipeline:
  validators:
    - image: gcr.io/kpt-fn/gatekeeper:unstable
```

In the constraint, we use `enforcementAction: warn` instead of
`enforcementAction: deny`.

```yaml
apiVersion: constraints.gatekeeper.sh/v1beta1
kind: K8sBannedConfigMapKeysV1
metadata:
  name: no-secrets-in-configmap
spec:
  enforcementAction: warn
  ...
```

### Function invocation

Get the package:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/gatekeeper/warning-only .
```

Run the function:

```shell
$ kpt fn render warning-only --results-dir=/tmp
```

### Expected result

Let's take a look at the structured results in `/tmp/results.yaml`:

```shell
apiVersion: kpt.dev/v1alpha2
kind: FunctionResultList
metadata:
  name: fnresults
exitCode: 0
items:
  - image: gcr.io/kpt-fn/gatekeeper:unstable
    exitCode: 0
    results:
      - message: |-
          The following banned keys are being used in the ConfigMap: {"private_key"}
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

To pass validation, let's replace the key `private_key` in the `ConfigMap` in
`resources.yaml` with something else e.g. `public_key`.
Rerun the command. It will no longer have the warning.

[gatekeeper]: https://catalog.kpt.dev/gatekeeper/v0.1/

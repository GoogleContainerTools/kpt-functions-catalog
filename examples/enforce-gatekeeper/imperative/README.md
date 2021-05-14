# enforce-gatekeeper: Imperative Example

### Overview

This example demonstrates how to imperatively invoke the [enforce-gatekeeper]
function to validate resources using gatekeeper constraints.

There are 3 resources: a ConstraintTemplate, a K8sBannedConfigMapKeysV1 and a
ConfigMap.
The constraint disallows using `private_key` as a key in the ConfigMap.

### Function invocation

Get the package:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/enforce-gatekeeper/imperative .
```

Create a directory for storing the structured output.

```shell
$ cd imperative
$ mkdir results
```

Run the function:

```shell
kpt fn eval --image=gcr.io/kpt-fn/enforce-gatekeeper:unstable --results-dir=results
```

### Expected result

You should see the output like the following:

```
[RUNNING] "gcr.io/kpt-fn/enforce-gatekeeper:unstable"
[FAIL] "gcr.io/kpt-fn/enforce-gatekeeper:unstable"
  Stderr:
    "The following banned keys are being used in the ConfigMap: {\"private_key\"}"
    "violatedConstraint: no-secrets-in-configmap"
  Exit Code: 1

For complete results, see results/results.yaml
```

Let's take a look at the structured result:

```yaml
apiVersion: kpt.dev/v1alpha2
kind: FunctionResultList
metadata:
  name: fnresults
exitCode: 1
items:
  - image: gcr.io/kpt-fn/enforce-gatekeeper:unstable
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

To pass validation, let's replace the key `private_key` in the ConfigMap in
`resources.yaml` with something else e.g. `public_key`.
Rerun the command. It will succeed.

[enforce-gatekeeper]: https://catalog.kpt.dev/enforce-gatekeeper/v0.1/

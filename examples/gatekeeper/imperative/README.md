# gatekeeper: Imperative Example

### Overview

This examples shows how to validate resources using gatekeeper constraints by
running [`gatekeeper`] function imperatively.

To ensure ConfigMaps do not contain fields with name `private_key`, we express
the constraints in the following resources in `policy.yaml`:

- `ConstraintTemplate`
- `K8sBannedConfigMapKeysV1`

We have a `ConfigMap` in `config-map.yaml` that violates the policy.

### Function invocation

Get the package:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/gatekeeper/imperative
```

Run the function with `--results-dir` flag:

```shell
$ kpt fn eval --image=gcr.io/kpt-fn/gatekeeper:unstable --results-dir=results
```

### Expected result

Let's take a look at the structured result:

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
`resources.yaml` with something else e.g. `public_key`. Rerun the command. It
will succeed.

[`gatekeeper`]: https://catalog.kpt.dev/gatekeeper/v0.1/

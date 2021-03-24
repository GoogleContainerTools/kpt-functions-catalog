# policy-controller-validate

### Overview

<!--mdtogo:Short-->

Validate the KRM resources using the policy controller.

<!--mdtogo-->


config-root/.../*-constraint.yaml and *-template.yaml define Policy Controller constraints and templates which all configs in config-root/ must pass.

### Synopsis

<!--mdtogo:Long-->
You can use the policy controller to validate KRM resources. To learn more about
the policy controller, see: https://cloud.google.com/anthos-config-management/docs/concepts/policy-controller.

The function takes 3 types of resources from the input resource list:

- constraint templates
- constraints
- KRM resources to be audited against

Every constraint should be backed by a constraint template that defines the
schema and logic of the constraint.

The function uses the constraints to audit the input KRM resources.

To learn more about how to write constraint templates and constraints, see:
https://cloud.google.com/anthos-config-management/docs/how-to/write-a-constraint-template
and
https://cloud.google.com/anthos-config-management/docs/how-to/creating-constraints.
<!--mdtogo-->

### Examples

<!--mdtogo:Examples-->
We have a constraint template which defines a rule that a config map can't have
any keys defined in the banned keys.

```yaml
apiVersion: templates.gatekeeper.sh/v1beta1
kind: ConstraintTemplate
metadata:
  name: k8sbannedconfigmapkeysv1
spec:
  crd:
    spec:
      names:
        kind: K8sBannedConfigMapKeysV1
        validation:
          openAPIV3Schema:
            properties:
              keys:
                type: array
                items:
                  type: string
  targets:
    - target: admission.k8s.gatekeeper.sh
      rego: |-
        package ban_keys

        violation[{"msg": sprintf("%v", [val])}] {
          keys = {key | input.review.object.data[key]}
          banned = {key | input.parameters.keys[_] = key}
          overlap = keys & banned
          count(overlap) > 0
          val := sprintf("The following banned keys are being used in the config map: %v", [overlap])
        }
```

We also have a constraint backed by the above constraint template. It defines
the target resources are the config maps, and the banned keys list only contains
"private-key".

```yaml
apiVersion: constraints.gatekeeper.sh/v1beta1
kind: K8sBannedConfigMapKeysV1
metadata:
  name: no-secrets-in-configmap
spec:
  match:
    kinds:
      - apiGroups:
          - ''
        kinds:
          - ConfigMap
  parameters:
    keys:
      - private_key
```

We have a config map to be audited.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: super-secret
  namespace: default
data:
  private_key: sensitive data goes here
```

To run the function to audit the resources:

```shell
kpt fn run gcr.io/kpt-fn/policy-controller-validate:unstable .
```

We will see the following validation error:
```
Found 1 violations:

[1] The following banned keys are being used in the config map: {"private_key"}

name: "super-secret"
path: resources.yaml
violatedConstraint: no-secrets-in-configmap

error: exit status 1
```

<!--mdtogo-->

// Code generated by "mdtogo"; DO NOT EDIT.
package generated

var GatekeeperShort = `The ` + "`" + `gatekeeper` + "`" + ` function wraps the [` + "`" + `Gatekeeper` + "`" + `] policy enforcement engine to
evaluate [Gatekeeper] constraints to validate the KRM resources.`
var GatekeeperLong = `
### Authoring Policies

The ` + "`" + `gatekeeper` + "`" + ` function follows the [executable configuration] pattern. There
are 2 kinds of resources needed to define a policy and they need to be provided
using ` + "`" + `input items` + "`" + ` along with other KRM resources to be validated.

- [Constraint Template]: Define the schema and logic of a policy. The policy
  logic in a Constraint Template must be written in the [Rego] language.
- [Constraint]: Signal the Gatekeeper the corresponding constraints need to be
  enforced. Every Constraint must be backed by a Constraint Template.

The constraint templates and the constraints must be provided using
` + "`" + `input items` + "`" + ` along with other KRM resources to be validated. Nothing need to be
provided in ` + "`" + `input functionConfig` + "`" + `.

The following is a ` + "`" + `ConstraintTemplate` + "`" + ` and a ` + "`" + `Constraint` + "`" + `:

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
            val := sprintf("The following banned keys are being used in the ConfigMap: %v", [overlap])
          }
  ---
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
`

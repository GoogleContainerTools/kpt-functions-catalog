# gatekeeper

### Overview

<!--mdtogo:Short-->

The `gatekeeper` function wraps the [`Gatekeeper`] policy enforcement engine to
evaluate [Gatekeeper] constraints to validate the KRM resources.

<!--mdtogo-->

It provides a very flexible and powerful way to validate KRM resources against
custom policies.

You can learn more about how to use the `Gatekeeper` project [here][howto]. The
concept document about the `OPA Constraint Framework` can be found
[here][concept].

<!--mdtogo:Long-->

### Authoring Policies

The `gatekeeper` function follows the [executable configuration] pattern. There
are 2 kinds of resources needed to define a policy and they need to be provided
using `input items` along with other KRM resources to be validated.

- [Constraint Template]: Define the schema and logic of a policy. The policy
  logic in a Constraint Template must be written in the [Rego] language.
- [Constraint]: Signal the Gatekeeper the corresponding constraints need to be
  enforced. Every Constraint must be backed by a Constraint Template.

The constraint templates and the constraints must be provided using
`input items` along with other KRM resources to be validated. Nothing need to be
provided in `input functionConfig`.

The following is a `ConstraintTemplate` and a `Constraint`:

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
```

<!--mdtogo-->

[Gatekeeper]: https://open-policy-agent.github.io/gatekeeper/website/docs/

[Constraint Template]: https://open-policy-agent.github.io/gatekeeper/website/docs/howto#constraint-templates

[Constraint]: https://open-policy-agent.github.io/gatekeeper/website/docs/howto#constraints

[Rego]: https://www.openpolicyagent.org/docs/latest/#rego

[howto]: https://open-policy-agent.github.io/gatekeeper/website/docs/howto

[concept]: https://github.com/open-policy-agent/frameworks/tree/master/constraint#opa-constraint-framework

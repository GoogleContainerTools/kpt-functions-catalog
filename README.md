# kpt functions catalog

This repository contains a catalog of kpt functions.

| Image                                     | Description                                                                                                                | Use Case       |
| ----------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- | -------------- |
| gcr.io/kpt-functions/source-yaml-dir      | Reads a directory of Kubernetes configuration recursively.                                                                 | Source         |
| gcr.io/kpt-functions/sink-yaml-dir        | Writes a directory of Kubernetes configuration. It maintains the original directory structure as read by source functions. | Sink           |
| gcr.io/kpt-functions/gatekeeper-validate  | Enforces OPA constraints on input objects. The constraints are also passed as part of the input to the function.           | Validation     |
| gcr.io/kpt-functions/recommend-psp        | [Demo] Mutates `PodSecurityPolicy` objects by setting `spec.allowPrivilegeEscalation` to `false`.                          | Recommendation |
| gcr.io/kpt-functions/validate-rolebinding | [Demo] Enforces a blacklist of `subjects` in `RoleBinding` objects.                                                        | Validation     |
| gcr.io/kpt-functions/hydrate-anthos-team  | [Demo] Reads custom resources of type `Team` and generates multiple `Namespace` and `RoleBinding` objects.                 | Generation     |
| gcr.io/kpt-functions/no-op                | [Demo] No Op function.                                                                                                     | Testing        |

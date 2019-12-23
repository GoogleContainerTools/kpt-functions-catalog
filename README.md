# KPT Functions catalog

This repository contains a catalog of KPT functions.

| Image                                     | Description                                                                                                                | Use Case       |
| ----------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- | -------------- |
| gcr.io/kpt-functions/read-yaml      | Reads a directory of Kubernetes configuration recursively.                                                                 | Source         |
| gcr.io/kpt-functions/write-yaml        | Writes a directory of Kubernetes configuration. It maintains the original directory structure as read by source functions. | Sink           |
| gcr.io/kpt-functions/gatekeeper-validate  | Enforces OPA constraints on input objects. The constraints are also passed as part of the input to the function.           | Compliance     |
| gcr.io/kpt-functions/mutate-psp        | [Demo] Mutates `PodSecurityPolicy` objects by setting `spec.allowPrivilegeEscalation` to `false`.                          | Recommendation |
| gcr.io/kpt-functions/validate-rolebinding | [Demo] Enforces a blacklist of `subjects` in `RoleBinding` objects.                                                        | Compliance     |
| gcr.io/kpt-functions/expand-team-cr  | [Demo] Reads custom resources of type `Team` and generates multiple `Namespace` and `RoleBinding` objects.                 | Generation     |
| gcr.io/kpt-functions/no-op                | [Demo] No Op function.                                                                                                     | Testing        |

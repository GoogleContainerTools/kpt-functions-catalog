# delete-annotations: Simple Example

### Overview

In this example, we will see how to delete annotations on a set of resources in a package/folder

### Fetch the example package

Get the example package by running the following commands:

```shell
$ kpt pkg get https://github.com/GoogleContainerTools/kpt-functions-catalog.git/examples/delete-annotations-simple
```

### Function invocation

Invoke the function by running the following command:

```shell
$ kpt fn render delete-annotations-simple
```

### Expected result

One of the two resources in `resources.yaml` should have been mutated with the annotation `annotation.to.delete` removed from `metadata.annotations`. There shouldn't be any changes to the second resource as it doesn't have the supplied annotation.

```yaml
apiVersion: constraints.gatekeeper.sh/v1beta1
kind: K8sRestrictRoleBindings
metadata:
  name: restrict-clusteradmin-rolebindings
  annotations:
    description: "Restricts use of the cluster-admin role."
    bundles.validator.forsetisecurity.org/cis-k8s-v1.5.1: 5.1.1
spec:
  enforcementAction: deny
  parameters:
    restrictedRole:
      apiGroup: "rbac.authorization.k8s.io"
      kind: "ClusterRole"
      name: "cluster-admin"
    allowedSubjects:
    - apiGroup: "rbac.authorization.k8s.io"
      kind: "Group"
      name: "system:masters"
---
apiVersion: constraints.gatekeeper.sh/v1beta1
kind: K8sProhibitRoleWildcardAccess
metadata:
  name: prohibit-role-wildcard-access
  annotations:
    description: "Restricts use of wildcards in Roles and ClusterRoles."
    bundles.validator.forsetisecurity.org/cis-k8s-v1.5.1: 5.1.3
spec:
  enforcementAction: deny

```

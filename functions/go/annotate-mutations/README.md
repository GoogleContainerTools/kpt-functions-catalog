# annotate-mutations

## Overview

<!--mdtogo:Short-->

The `annotate-mutations` function reads `apply-time-mutation` comments on resource YAML and adds
the equivalent `config.k8s.io/apply-time-mutation` to the resource.

<!--mdtogo-->

<!--mdtogo:Long-->

## Usage

annotate-mutations function is expected to be executed imperatively like:

```shell
kpt fn eval --include-meta-resources --image gcr.io/kpt-fn/annotate-mutations:unstable
```

The `annotate-mutations` function does the following:

1.  Scans the package for `apply-time-mutation` comment markup.
2.  Appends the equivalent `config.k8s.io/apply-time-mutation` annotation to the same.

<!--mdtogo-->

## Examples

<!--mdtogo:Examples-->

Appending an `config.k8s.io/apply-time-mutation` annotation based on a comment.

Let's start with the Kptfile of an example package.

```yaml
apiVersion: iam.cnrm.cloud.google.com/v1beta1
kind: IAMPolicyMember
metadata:
  name: my-policy
  namespace: example-namespace
spec:
  member: placeholder # apply-time-mutation: "serviceAccount:service-${resourcemanager.cnrm.cloud.google.com/namespaces/example-namespace/Project/example-name:$.status.number}@container-engine-robot.iam.gserviceaccount.com"
```

Invoke the function:

```shell
kpt fn eval --include-meta-resources --image gcr.io/kpt-fn/annotate-mutations:unstable
```

Resource will be updated to the following:

```yaml
apiVersion: iam.cnrm.cloud.google.com/v1beta1
kind: IAMPolicyMember
metadata:
  name: my-policy
  namespace: example-namespace
  annotations:
    config.k8s.io/apply-time-mutation: |
      - sourceRef:
          group: resourcemanager.cnrm.cloud.google.com
          kind: Project
          name: example-name
          namespace: example-namespace
        sourcePath: $.status.number
        targetPath: $.spec.member
        token: ${.status.number}
spec:
  member: serviceAccount:service-${.status.number}@container-engine-robot.iam.gserviceaccount.com
```

<!--mdtogo-->

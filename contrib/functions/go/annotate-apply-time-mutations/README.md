# annotate-apply-time-mutations

## Overview

<!--mdtogo:Short-->

The `annotate-apply-time-mutations` function enables authors to use inline
comments OR custom resource objects to generate
[apply-time-mutation annotations](https://kpt.dev/reference/annotations/apply-time-mutation/).

This can help simplify the authoring of apply-time-mutation annotations, without
needing to directly generate, manipulate, or template the YAML strings used by
those annotations.

<!--mdtogo-->

<!--mdtogo:Long-->

## Usage

The function can be executed declaratively using `kpt fn render`.

To configure this, add the function to the pipeline config in the Kptfile:

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
pipeline:
  mutators:
    - image: gcr.io/kpt-fn-contrib/annotate-apply-time-mutations:unstable
```

The function can also be executed imperatively:

```shell
kpt fn eval --image gcr.io/kpt-fn-contrib/annotate-apply-time-mutations:unstable
```

The function will generate and write the
`config.kubernetes.io/apply-time-mutation` annotation on the target object, the
object with the apply-time-mutation comment on one or more of its fields.

The function does not perform the mutation on the target object itself. The
mutation is performed by `kpt live apply`, which reads the annotation as input.
So the function should be run by the user before applying.

## Inline Field Comment Input

Inline field comments can be used as an alternate way to specify mutations. This
function will convert `apply-time-mutation` comments into apply-time-mutation
annotations.

With inline comments, the mutation is specified closer to the target field that
will be updated. This can aid debugging and onboarding by reducing indireciton.
It also reduces the configuration required, because the target object and field
don't need to be explicitly specified.

Inline field comments can be specified with the following format:

```
# apply-time-mutation: [prefix]${[group]/[version]/namespaces/[source-namespace]/[kind]/[source-name]:[source-field-path]}[suffix]
```

`prefix` and `suffix` are optional. They are constant strings that surround dynamic
content from the source field. When specified, the function will copy them into
the field itself, surrounding a generated token.

```
field: [prefix]$ref[suffix] # apply-time-mutation: ...
```

`version` is also optional. It's generally recommended to avoid specifying the
version. This allows the object reference to apply to any version of the
resource API, which makes the reference less brittle and can survive CRD version
updates. If you specify the version, only that API version will match when
looking up the source object.

## Custom Resource Object Input

Custom resource objects can be used to specify mutations. This function will
convert `ApplyTimeMutation` objects into apply-time-mutation annotations.

With custom resource objects, the mutation is specified with KRM, which is
more indirect, but allows for generation, manipulation, and templating of the
mutation specification. One big win with this method is that kpt setters can be
used to configure the source and target object references (ex: name & namespace).

`ApplyTimeMutation` resource objects can be specified with the following format:

```yaml
apiVersion: function.kpt.dev/v1alpha1
kind: ApplyTimeMutation
metadata:
  name: example
  annotations:
    config.kubernetes.io/local-config: "true"
spec:
  targetRef:
    kind: ConfigMap
    name: target-object
    namespace: test-namespace
  substitutions:
  - sourceRef:
      kind: ConfigMap
      name: source-object
      namespace: test-namespace
    sourcePath: $.spec.data
    targetPath: $.spec.data
```

The `ApplyTimeMutation` resource follows the standard
[Kubernetes Resource Model (KRM)](https://github.com/kubernetes/design-proposals-archive/blob/main/architecture/resource-management.md)
with top level `apiVersion`, `kind`, `metadata` fields, as well as the
conventional `spec` field for specification configuration. Like other KRM
resources, the `ApplyTimeMutation` resource also supports the standard metadata
fields, like label and annotation. The function will simply ignore them.

If you're familiar with the apply-time-mutation annotation syntax, the
`spec.substitutions` field of the `ApplyTimeMutation` resource should look
familiar. For details about the substitution schema, see the
[apply-time-mutation reference docs](https://kpt.dev/reference/annotations/apply-time-mutation/).

In addition to the substitutions, when using the `ApplyTimeMutation` resource,
the target object must be referenced. The `spec.targetRef` field uses the
[ObjectReference schema](https://kpt.dev/reference/annotations/apply-time-mutation/?id=objectreference).
The target object reference specifies which object will receive the
apply-time-mutation annotation, the object with target fields to be modified.

Remember to use the
[local-config annotation](https://kpt.dev/reference/annotations/local-config/)
so the resource is not applied by `kpt live apply`.

<!--mdtogo-->

## Examples

<!--mdtogo:Examples-->

### Inline Field Comment Example

This example demonstrates generating an `config.kubernetes.io/apply-time-mutation`
annotation based on an `apply-time-mutation` comment.

Start with a source object:

```yaml
apiVersion: resourcemanager.cnrm.cloud.google.com/v1beta1
kind: Project
metadata:
  name: my-project
  namespace: example-namespace
spec:
  name: My Project
  organizationRef:
    external: "123456789012"
  billingAccountRef:
    external: "AAAAAA-BBBBBB-CCCCCC"
# The status will be populated by a controller after the object is created
# status:
#   number: "1234567890123"
```

Add an `apply-time-mutation` comment to the target object:

```yaml
apiVersion: iam.cnrm.cloud.google.com/v1beta1
kind: IAMPolicyMember
metadata:
  name: my-policy
  namespace: example-namespace
spec:
  member: placeholder # apply-time-mutation: "serviceAccount:service-${resourcemanager.cnrm.cloud.google.com/namespaces/example-namespace/Project/my-project:$.status.number}@container-engine-robot.iam.gserviceaccount.com"
```

Invoke the function:

```shell
kpt fn eval --image gcr.io/kpt-fn-contrib/annotate-apply-time-mutations:unstable
```

The target object will be updated to the following:

```yaml
apiVersion: iam.cnrm.cloud.google.com/v1beta1
kind: IAMPolicyMember
metadata:
  name: my-policy
  namespace: example-namespace
  annotations:
    config.kubernetes.io/apply-time-mutation: |
      - sourceRef:
          group: resourcemanager.cnrm.cloud.google.com
          kind: Project
          name: my-project
          namespace: example-namespace
        sourcePath: $.status.number
        targetPath: $.spec.member
        token: ${ref0}
spec:
  member: serviceAccount:service-${ref0}@container-engine-robot.iam.gserviceaccount.com # apply-time-mutation: ...
```

### Custom Resource Object Example

This example demonstrates generating an `config.kubernetes.io/apply-time-mutation`
annotation based on an `ApplyTimeMutation` object.

Start with source and target objects:

```yaml
apiVersion: resourcemanager.cnrm.cloud.google.com/v1beta1
kind: Project
metadata:
  name: my-project
  namespace: example-namespace
spec:
  name: My Project
  organizationRef:
    external: "123456789012"
  billingAccountRef:
    external: "AAAAAA-BBBBBB-CCCCCC"
# The status will be populated by a controller after the object is created
# status:
#   number: "1234567890123"
```

```yaml
apiVersion: iam.cnrm.cloud.google.com/v1beta1
kind: IAMPolicyMember
metadata:
  name: my-policy
  namespace: example-namespace
spec:
  member: serviceAccount:service-${project-number}@container-engine-robot.iam.gserviceaccount.com
```

Specify the mutation with an `ApplyTimeMutation` object:

```yaml
apiVersion: function.kpt.dev/v1alpha1
kind: ApplyTimeMutation
metadata:
  name: example
  annotations:
    config.kubernetes.io/local-config: "true"
spec:
  targetRef:
    group: iam.cnrm.cloud.google.com
    kind: IAMPolicyMember
    name: my-policy
    namespace: example-namespace
  substitutions:
  - sourceRef:
      group: resourcemanager.cnrm.cloud.google.com
      kind: Project
      name: my-project
      namespace: example-namespace
    sourcePath: $.status.number
    targetPath: $.spec.member
    token: "${project-number}"
```

The target object will be updated to the following:

```yaml
apiVersion: iam.cnrm.cloud.google.com/v1beta1
kind: IAMPolicyMember
metadata:
  name: my-policy
  namespace: example-namespace
  annotations:
    config.kubernetes.io/apply-time-mutation: |
      - sourceRef:
          group: resourcemanager.cnrm.cloud.google.com
          kind: Project
          name: my-project
          namespace: example-namespace
        sourcePath: $.status.number
        targetPath: $.spec.member
        token: ${project-number}
spec:
  member: serviceAccount:service-${project-number}@container-engine-robot.iam.gserviceaccount.com
```

<!--mdtogo-->

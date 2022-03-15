# annotate-apply-time-mutations

## Overview

<!--mdtogo:Short-->

The `annotate-apply-time-mutations` function enables authors to use alternate
input methods to generate
[apply-time-mutation annotations](https://kpt.dev/reference/annotations/apply-time-mutation/).

Input formats:
- [Inline field comment: apply-time-mutation](#inline-field-comment-input)
- [Custom resource object: ApplyTimeMutation](#custom-resource-object-input)

This can help simplify the authoring of apply-time-mutation annotations, without
needing to directly generate, manipulate, or template the YAML strings used by
those annotations.

<!--mdtogo-->

<!--mdtogo:Long-->

## Usage

The `annotate-apply-time-mutations` function can be executed by itself or as
part of a [kpt workflow](https://kpt.dev/book/02-concepts/02-workflows).

To execute by itself:

```shell
kpt fn eval --image gcr.io/kpt-fn-contrib/annotate-apply-time-mutations:v0.1.0
```

To execute as part of a kpt workflow, first modify the Kptfile to add the
function to the pipeline:

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
pipeline:
  mutators:
    - image: gcr.io/kpt-fn-contrib/annotate-apply-time-mutations:v0.1.0
```

Then execute the pipeline:

```
kpt fn render
```

Either way, the function will read the input files and generate 
`config.kubernetes.io/apply-time-mutation` annotations on the target object(s).

The `annotate-apply-time-mutations` function does not perform the mutation on
the target object itself. The mutation is performed by `kpt live apply`, which
reads the annotation as input. So the function needs to be run by the user
before applying.

### Inline Field Comment Input

Inline field comments can be used as an alternate way to specify mutations. This
function will convert `apply-time-mutation` comments into apply-time-mutation
annotations.

With inline comments, the mutation is specified closer to the target field that
will be updated. This can aid debugging and onboarding by reducing indireciton.
It also reduces the configuration required, because the target object and field
don't need to be explicitly specified.

Inline field comments can be specified with the following format:

```yaml
field: "" # apply-time-mutation: [PREFIX]${GROUP/[VERSION/][namespaces/NAMESPACE/]KIND/NAME:FIELD_PATH}[SUFFIX]
```

Fields and delimiters surrounded in square brackets (`[]`) are optional. Your
comment should include the curly braces (`${}`) but NOT the square brackets.

- `PREFIX` (Optional) - A string to prepend to the substituted value.
- `GROUP` - The API group of the source object. For "core" resources, the group
  is the empty string, with the trailing slash (`/`) delimiter retained.
- `VERSION` (Optional) - The API version of the source object. When supplied, it
  will match only objects using this exact API version. It's recommended to
  just use `GROUP` without version to make the reference less brittle and able
  to survive CRD version updates.
- `NAMESPACE` (Optional) - The namespace of the source object, required for 
  namespace-scoped resources
- `KIND` - The kind of the source object
- `NAME` - The name of the source object
- `FIELD_PATH` - A JSONPath expression that identifies the source object field
- `SUFFIX` (Optional) - A string to append to the substituted value

When the function runs, if `PREFIX` or `SUFFIX` is specified, the field value
will be replaced with a string including the `PREFIX` and `SUFFIX`, surrounding
a generated token for substitution.

```yaml
field: "PREFIX${ref1}SUFFIX" # apply-time-mutation: ...
```

When the function runs, if neither `PREFIX` nor `SUFFIX` are specified, the
field value will not be replaced and no token will be specified, causing the
whole field value to be replaced, using the type of the source object field.

The apply-time-mutation comment will be preserved so that the function is
idempotent, producing the same output when run multiple times.

### Custom Resource Object Input

Custom resource objects can be used to specify mutations. This function will
convert `ApplyTimeMutation` objects into apply-time-mutation annotations.

With custom resource objects, the mutation is specified with KRM, which is
more indirect, but allows for generation, manipulation, and templating of the
mutation specification. One big win with this method is that kpt setters can be
used to configure the source and target object references (ex: name & namespace).

`ApplyTimeMutation` resource objects can be specified with the following format:

```yaml
apiVersion: fn.kpt.dev/v1alpha1
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
kpt fn eval --image gcr.io/kpt-fn-contrib/annotate-apply-time-mutations:v0.1.0
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
apiVersion: fn.kpt.dev/v1alpha1
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

Invoke the function:

```shell
kpt fn eval --image gcr.io/kpt-fn-contrib/annotate-apply-time-mutations:v0.1.0
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

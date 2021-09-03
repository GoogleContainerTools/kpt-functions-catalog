# annotate-apply-time-mutations

## Overview

<!--mdtogo:Short-->

The `annotate-apply-time-mutations` function enables authors to use inline comments,
rather than annotations, to specify field replacements using the apply time mutation feature.

It works by reading `# apply-time-mutation:` comments on resource YAML and adds the equivalent
`config.kubernetes.io/apply-time-mutation` annotation to the resource.

The `config.kubernetes.io/apply-time-mutation` annotation is read by the [apply time mutation]
functionality which patches the resource config at apply time, during `kpt live apply`, with
the referenced resource's live value.

<!--mdtogo-->

<!--mdtogo:Long-->

## Usage

The annotate-apply-time-mutations function can be executed declaratively as part of `kpt fn render`

```yaml
apiVersion: kpt.dev/v1
kind: Kptfile
pipeline:
  mutators:
    - image: gcr.io/kpt-fn-contrib/annotate-apply-time-mutations:unstable
```

or imperatively like:

```shell
kpt fn eval --image gcr.io/kpt-fn-contrib/annotate-apply-time-mutations:unstable
```


The `annotate-apply-time-mutations` function does the following:

1.  Scans the package for `apply-time-mutation` comment markup.
2.  Appends the equivalent `config.k8s.io/apply-time-mutation` annotation to the same.

The expected `apply-time-mutation` comment format is:

`# apply-time-mutation: [prefix]${[group]/[version]/namespaces/[source namespace]/[kind]/[source name]:[source field path]}[suffix]`

Prefix, version, and suffix are optional fields.

For fields with a substitution as well as a constant prefix and/or suffix, this function will insert a replacement token in the field, matched in the annotation.
`field: [prefix]replaceme[suffix] # apply-time-mutation: ...`

<!--mdtogo-->

## Examples

<!--mdtogo:Examples-->

Appending an `config.kubernetes.io/apply-time-mutation` annotation based on a comment.

Let's start with a sample resource.

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
kpt fn eval --image gcr.io/kpt-fn-contrib/annotate-apply-time-mutations:unstable
```

Resource will be updated to the following:

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
          name: example-name
          namespace: example-namespace
        sourcePath: $.status.number
        targetPath: $.spec.member
        token: $ref0
spec:
  member: serviceAccount:service-$ref0@container-engine-robot.iam.gserviceaccount.com # apply-time-mutation: ...
```

<!--mdtogo-->

[apply time mutation] https://kpt.dev/reference/cli/live/apply/

package consts

// annotationFieldSpecs update the KRM resources' annotations whose field spec matches the path.
// This requires the annotation to already exist in the KRM resource, which is different from kustomize's
// commonAnnotations that creates the annotation if not exist.
const AnnotationFieldSpecs = `
annotations:
- path: metadata/annotations

- path: spec/template/metadata/annotations
  version: v1
  kind: ReplicationController

- path: spec/template/metadata/annotations
  kind: Deployment

- path: spec/template/metadata/annotations
  kind: ReplicaSet

- path: spec/template/metadata/annotations
  kind: DaemonSet

- path: spec/template/metadata/annotations
  kind: StatefulSet

- path: spec/template/metadata/annotations
  group: batch
  kind: Job

- path: spec/jobTemplate/metadata/annotations
  group: batch
  kind: CronJob

- path: spec/jobTemplate/spec/template/metadata/annotations
  group: batch
  kind: CronJob

`

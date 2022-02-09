package consts

// ProjectFieldSpecs contains the fieldSpec paths of Google Cloud ProjectID
const ProjectFieldSpecs = `
projectFieldSpec:
# Blueprint redis-bucket
- path: spec/authorizedNetworkRef/external 
  regexPattern: (?P<prefix>projects\/)(?P<projectID>\S+)(?P<suffix>\/global\/networks\/default)
  group: redis.cnrm.cloud.google.com
  version: v1beta1
  kind: RedisInstance

# Blueprint log export
- path: spec/resourceRef/external
  regexPattern: (?P<prefix>projects\/)(?P<projectID>\S+)
  group: iam.cnrm.cloud.google.com
  version: v1beta1
  kind: IAMPolicyMember
`

/*
# Test-Only
- path: metadata/name
version: v1
group: apps
kind: Deployment
create: true
*/

/* UNSUPPORTED Project FieldSpec due to substitution/partial-setter
- path: metadata/annotations[cnrm.cloud.google.com/project-id]
  group: storage.cnrm.cloud.google.com
  version: v1beta1
  kind: StorageBucket
  create: true

- path: spec/workloadIdentityConfig/identityNamespace
  regexPattern: "(\s+)\.svc\.id\.goog"
  group: container.cnrm.cloud.google.com
  version: v1beta1
  kind: ContainerCluster
  create: true
- path: spec/networkRef/external
  regexPattern: "projects/(\s+)/global/networks/default""
  group: container.cnrm.cloud.google.com
  version: v1beta1
  kind: ContainerCluster
  create: true

- path: spec/bindings[]/members[]/member
  regexPattern: "^serviceAccount:\s+@(\s+).iam.gserviceaccount.com"
  group: iam.cnrm.cloud.google.com
  version: v1beta1
  kind: IAMPartialPolicy
  create: true
- path: spec/authority/issuer
  regexPattern: "https://container.googleapis.com/v1/projects/(\s+)/locations/\s+/clusters/\s+"
  group: gkehub.cnrm.cloud.google.com
  version: v1beta1
  kind: GKEHubMembership
- path: spec/endpoint/gkeCluster/resourceRef/external
  regexPattern: "//container.googleapis.com/projects/(\s+)/locations/\s+/clusters/\s+"
  group: gkehub.cnrm.cloud.google.com
  version: v1beta1
  kind: GKEHubMembership
  create: true
- path: spec/projectRef/external
  regexPattern: "//container.googleapis.com/projects/(\s+)/locations/\s+/clusters/\s+"
  group: gkehub.cnrm.cloud.google.com
  version: v1beta1
  kind: GKEHubFeature
  create: true
- path: spec/projectRef/configmanagement/configSync/git/gcpServiceAccountRef/external
  regexPattern: "\s+@(\s+).iam.gserviceaccount.com"
  group: gkehub.cnrm.cloud.google.com
  version: v1beta1
  kind: GKEHubFeatureMembership
  create: true
*/

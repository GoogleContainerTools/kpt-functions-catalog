package fieldspec

/* CustomNameFieldSpecs list should be adjusted actively before we get a variant constructor solution.
 */

const (
	CustomNameFieldSpecs = `
customMetaName:
- path: metadata/name 
  group: storage.cnrm.cloud.google.com

- path: metadata/name 
  group: serviceusage.cnrm.cloud.google.com

- path: metadata/name 
  group: redis.cnrm.cloud.google.com

- path: metadata/name 
  group: spanner.cnrm.cloud.google.com

- path: metadata/name
  kind: Service
  group: serviceusage.cnrm.cloud.google.com

- path: metadata/name
  kind: LoggingLogSink
  group: logging.cnrm.cloud.google.com

- path: metadata/name
  kind: IAMPolicyMember
  group: iam.cnrm.cloud.google.com

- path: metadata/name
  kind: PubSubTopic
  group: pubsub.cnrm.cloud.google.com

- path: metadata/name
  kind: BigQueryDataset
  group: bigquery.cnrm.cloud.google.com

- path: annotations/cnrm.cloud.google.com/project-id
  kind: SetAnnotations
  group: fn.kpt.dev
  version: v1alpha1


# A fieldSpec object under customMetaName. 
# - path: <fieldspec>
#   group: <API Group Name> if ignored, matches all 
#   version: <API Version> if ignored, matches all
#   kind: <Kind> if ignored, matches all
#   create: [true|false] default to false, if set to true, create the field path in resource. 
`
)
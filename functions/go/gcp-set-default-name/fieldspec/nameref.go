package fieldspec

const (
	NameReferenceFieldSpecs = `
nameReference:
# Blueprint redis-bucket
- kind: RedisInstance
  group: redis.cnrm.cloud.google.com
  version: v1beta1
  fieldSpecs:
  - path: spec/displayName
    kind: RedisInstance
    group: redis.cnrm.cloud.google.com
    version: v1beta1

# Blueprint spanner
- kind: SpannerInstance
  group: spanner.cnrm.cloud.google.com
  version: v1beta1
  fieldSpecs:
  - path: spec/displayName
    kind: SpannerInstance
    group: spanner.cnrm.cloud.google.com
    version: v1beta1
  - path: spec/instanceRef/name
    kind: SpannerDatabase
    group: spanner.cnrm.cloud.google.com
    version: v1beta1

# Blueprint log-export
- kind: LoggingLogSink
  group: logging.cnrm.cloud.google.com
  version: v1beta1
  fieldSpecs:
  - path: spec/memberFrom/logSinkRef/name
    kind: IAMPolicyMember
    group: iam.cnrm.cloud.google.com
    version: v1beta1
- kind: PubSubTopic
  group: pubsub.cnrm.cloud.google.com
  version: v1beta1
  fieldSpecs:
  - path: spec/destination/pubSubTopicRef/name
    kind: LoggingLogSink
    group: logging.cnrm.cloud.google.com
    version: v1beta1
  - path: spec/memberFrom/resourceRef/name
    kind: IAMPolicyMember
    group: aim.cnrm.cloud.google.com
    version: v1beta1

- kind: BigQueryDataset
  group: bigquery.cnrm.cloud.google.com
  version: v1beta1
  fieldSpecs:
  - path: spec/destination/bigQueryDatasetRef/name
    kind: LoggingLogSink
    group: logging.cnrm.cloud.google.com
    version: v1beta1
`
)

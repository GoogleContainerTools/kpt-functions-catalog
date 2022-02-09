package filedspec

const LocationFieldSpecs = `
regions:
# Storage and BigQueryDataset instances.
- path: spec/location
  version: v1beta1

- path: spec/region
  group: redis.cnrm.cloud.google.com
  version: v1beta1
  kind: RedisInstance

- path: spec/config
  group: spanner.cnrm.cloud.google.com
  version: v1beta1
  kind: SpannerInstance
  regexPattern: (?P<prefix>regional-)(?P<location>\S+)

zones:
- path: spec/location
  apiVersion: test.cnrm.cloud.google.com/v1beta1
  kind: Test
  create: false
`

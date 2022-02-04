package fieldspec

const (
	RegionFieldSpecs = `
region:
- path: spec/location
  group: storage.cnrm.cloud.google.com
  version: v1beta1
  kind: StorageBucket

- path: spec/region
  group: redis.cnrm.cloud.google.com
  version: v1beta1
  kind: RedisInstance
`
)

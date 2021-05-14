# kubeval

### Overview

Use [kubeval] to validate KRM resources against their [json schemas].

### Synopsis

The function configuration must be a ConfigMap.

The following keys can be used in the `data` field of the ConfigMap, and all of
them are optional:

- `schema_location`: The base URI used to fetch the json schemas. The default
  is empty. This feature only works with imperative runs, since declarative runs
  allow neither network access nor volume mount.
- `additional_schema_locations`: List of secondary base URIs used to fetch the
  json schemas.  These URIs will be used if the URI specified by
  `schema_location` did not have the required schema.  The default is empty.
  This feature only works with imperative runs.
- `ignore_missing_schemas`: Skip validation for resources without a schema. The
  default is `false`.
- `skip_kinds`: Comma-separated list of case-sensitive kinds to skip when
  validating against schemas. The default is empty.
- `strict`: Disallow additional properties that are not in the schemas. The
  default is `false`.

The following is an example function configuration:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-func-config
data:
  schema_location: "file:///abs/path/to/your/schema/directory"
  additional_schema_locations: "https://kubernetesjsonschema.dev,file:///abs/path/to/your/other/schema/directory"
  ignore_missing_schemas: "false"
  skip_kinds: "DaemonSet,MyCRD"
  strict: "true"
```

If neither `schema_location` nor `additional_schema_locations` is provided, we
will convert the baked-in OpenAPI document to json schemas and use them.
The baked-in OpenAPI document is from a GKE cluster with version v1.19.8. The
OpenAPI document contains kubernetes built-in types and some GCP CRDs (e.g.
BackendConfig), but it currently doesn't contain Config Connector CRDs.

[kubeval]:https://kubeval.com
[json schemas]:https://json-schema.org

# kubeval

### Overview

Use [kubeval] to validate KRM resources against their [json schemas].

### Synopsis

The function configuration must be a ConfigMap.

The following keys can be used in the `data` field of the ConfigMap, and all of
them are optional:

- `schema_location`: The base URL used to fetch the json schemas.
- `additional_schema_locations`: List of secondary base URLs used to fetch the
  json schemas.  These URLs will be used if the URL specified by
  `schema_location` did not have the required schema.
- `ignore_missing_schemas`: Skip validation for resources without a schema.  If
  omitted, a default value of false will be assumed.
- `skip_kinds`: Comma-separated list of case-sensitive kinds to skip when
  validating against schemas.  If omitted, no kinds will be skipped.
- `strict`: Disallow additional properties that are not in the schemas.  If
  omitted, a default value of false will be assumed.

If neither `schema_location` nor `additional_schema_locations` is provided, we
will convert the baked-in OpenAPI document to json schemas and use them.

Note: `kpt fn render` allow neither network access nor volume mount. That means
you need to use the baked-in OpenAPI schema when using this function in
`kpt fn render`.

[kubeval]:https://kubeval.com
[json schemas]:https://json-schema.org
# kubeval

### Overview

Validate KRM resources using [kubeval].

### Synopsis

`kubeval` allows you to validate KRM resources against their [json schemas].

The json schemas can be provided via `schema_location` and
`additional_schema_locations`. If neither `schema_location` nor
`additional_schema_locations` is provided, we will convert the baked-in OpenAPI
document to json schemas and use them.

This function can be configured using a ConfigMap with the following keys, all
of which are optional:

- `schema_location`: The base URL used to download schemas.
- `additional_schema_locations`: List of secondary base URLs used to download
schemas.  These URLs will be used if the URL specified by schema_location did
not have the required schema.  By default, there are no secondary URLs, and only
the primary base URL will be used.
- `ignore_missing_schemas`: Skip validation for resources without a schema.  If
omitted, a default value of false will be assumed.
- `skip_kinds`: Comma-separated list of case-sensitive kinds to skip when
validating against schemas.  If omitted, no kinds will be skipped.
- `strict`: Disallow additional properties that are not in the schemas.  If
omitted, a default value of false will be assumed.

[kubeval]:https://kubeval.com
[json schemas]:https://json-schema.org
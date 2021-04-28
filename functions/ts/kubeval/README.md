# kubeval

### Overview

Validates configuration using [kubeval].

### Synopsis

If neither `schema_location` nor `additional_schema_locations` is provided, we
will use the baked-in schemas. Otherwise, we will try to download the schemas.

Configured using a ConfigMap with the following keys, all of which are optional:

- `schema_location`: The base URL used to download schemas.
- `additional_schema_locations`: List of secondary base URLs used to download
schemas.  These URLs will be used if the URL specified by schema_location
did not have the required schema.  By default, there are no secondary URLs,
and only the primary base URL will be used.
- `ignore_missing_schemas`: Skip validation for resource definitions without a
schema.  If omitted, a default value of false will be assumed.
- `skip_kinds`: Comma-separated list of case-sensitive kinds to skip when
validating against schemas.  If omitted, no kinds will be skipped.
- `strict`: Disallow additional properties not in schema.  If omitted, a default
value of false will be assumed.

[kubeval]:https://kubeval.com

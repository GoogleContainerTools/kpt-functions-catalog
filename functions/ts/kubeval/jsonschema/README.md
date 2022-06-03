# Generate json schema from openapi

We use [openapi2jsonschema](https://github.com/instrumenta/openapi2jsonschema) to convert OpenAPI to json schema files to be consumed by `kubeval`.

```shell
cd functions/ts/kubeval/jsonschema
```

We are going to generate 2 sets of json schema files with and without `--strict` flag.

```shell
python3 ../third_party/github.com/instrumenta/openapi2jsonschema/openapi2jsonschema/command.py --kubernetes --expanded --stand-alone --strict -o master-standalone-strict ../openapi/openapi.json
```

```shell
python3 ../third_party/github.com/instrumenta/openapi2jsonschema/openapi2jsonschema/command.py --kubernetes --expanded --stand-alone -o master-standalone ../openapi/openapi.json
```

The generated schema will be used as the default built-in schema.

import {
  Configs,
  KubernetesObject,
  kubernetesObjectResult,
  generalResult,
  Result,
} from 'kpt-functions';
import { ChildProcess, spawn, spawnSync } from 'child_process';
import { Writable } from 'stream';

const DEFAULT_SCHEMA_LOCATION = '/tmp';
const DEFAULT_OPENAPI_LOCATION = '/home/node/openapi.json';

const SCHEMA_LOCATION = 'schema_location';
const ADDITIONAL_SCHEMA_LOCATIONS = 'additional_schema_locations';
const IGNORE_MISSING_SCHEMAS = 'ignore_missing_schemas';
const SKIP_KINDS = 'skip_kinds';
const STRICT = 'strict';

type Feedback = FeedbackItem[];

interface FeedbackItem {
  filename: string;
  kind: string;
  status: 'valid' | 'invalid';
  errors: string[];
}

export async function kubeval(configs: Configs): Promise<void> {
  const schemaLocation = configs.getFunctionConfigValue(SCHEMA_LOCATION);
  const additionalSchemaLocationsStr = configs.getFunctionConfigValue(
    ADDITIONAL_SCHEMA_LOCATIONS
  );
  const additionalSchemaLocations = additionalSchemaLocationsStr
    ? additionalSchemaLocationsStr.split(',')
    : [];
  const ignoreMissingSchemas = JSON.parse(
    configs.getFunctionConfigValue(IGNORE_MISSING_SCHEMAS) || 'false'
  );
  const skipKindsStr = configs.getFunctionConfigValue(SKIP_KINDS);
  const skipKinds = skipKindsStr ? skipKindsStr.split(',') : [];
  const strict = JSON.parse(configs.getFunctionConfigValue(STRICT) || 'false');

  const results: Result[] = [];

  // Convert openapi to json schema if neither schema_location nor
  // additional_schema_locations is provided.
  if (!schemaLocation && additionalSchemaLocations.length === 0) {
    await runOpenapi2jsonschema(configs, strict, results);
  }

  const args = buildKubevalArgs(
    schemaLocation,
    additionalSchemaLocations,
    ignoreMissingSchemas,
    skipKinds,
    strict
  );

  for (const object of configs.getAll()) {
    await runKubeval(object, results, args);
  }

  configs.addResults(...results);
}

async function runOpenapi2jsonschema(
  configs: Configs,
  strict: boolean,
  results: Result[]
): Promise<void> {
  const apiVersionKindSet = new Set();
  for (const object of configs.getAll()) {
    const avk = object.apiVersion + ',' + object.kind;
    if (!apiVersionKindSet.has(avk)) {
      apiVersionKindSet.add(avk);
    }
  }
  if (apiVersionKindSet.size > 0) {
    const openapi2jsonschemaArgs = [
      '--kubernetes',
      '--expanded',
      '--stand-alone',
      '--apiversionkind',
      Array.from(apiVersionKindSet).join(';'),
    ];
    if (strict) {
      openapi2jsonschemaArgs.push('--strict');
      openapi2jsonschemaArgs.push(
        '-o',
        DEFAULT_SCHEMA_LOCATION + '/master-standalone-strict'
      );
    } else {
      openapi2jsonschemaArgs.push(
        '-o',
        DEFAULT_SCHEMA_LOCATION + '/master-standalone'
      );
    }
    openapi2jsonschemaArgs.push(DEFAULT_OPENAPI_LOCATION);

    const openapi2jsonschemaProcess = spawnSync(
      'openapi2jsonschema',
      openapi2jsonschemaArgs,
      {
        encoding: 'utf-8',
        stdio: [process.stdin, 'pipe', 'pipe'],
      }
    );
    if (openapi2jsonschemaProcess.status !== 0) {
      const result = generalResult(
        String(openapi2jsonschemaProcess.stdout) +
          String(openapi2jsonschemaProcess.stderr),
        'error',
        undefined
      );
      results.push(result);
    }
  }
}

async function runKubeval(
  object: KubernetesObject,
  results: Result[],
  args: string[]
): Promise<void> {
  const kubevalProcess = spawn('kubeval', args, {
    stdio: ['pipe', 'pipe', process.stderr],
  });
  const serializedObject = JSON.stringify(object);
  await writeToStream(kubevalProcess.stdin, serializedObject);
  kubevalProcess.stdin.end();
  const rawOutput = await readStdoutToString(kubevalProcess);
  try {
    const feedback = JSON.parse(rawOutput) as Feedback;

    for (const { status, errors } of feedback) {
      if (status !== 'valid') {
        for (const error of errors) {
          const [path, ...rest] = error.split(':');
          let result;
          if (rest.length > 0) {
            result = kubernetesObjectResult(
              rest.join(':').trim(),
              object,
              {
                path,
              },
              'error'
            );
          } else {
            result = kubernetesObjectResult(error, object, undefined, 'error');
          }
          results.push(result);
        }
      }
    }
  } catch (error) {
    results.push(
      kubernetesObjectResult(
        'Failed to parse raw kubeval output:\n' +
          error.message +
          '\n\n' +
          rawOutput,
        object
      )
    );
  }
}

function buildKubevalArgs(
  schemaLocation: string | undefined,
  additionalSchemaLocations: string[],
  ignoreMissingSchemas: boolean,
  skipKinds: string[],
  strict: boolean
) {
  const args = ['--output', 'json'];

  if (schemaLocation) {
    args.push('--schema-location');
    args.push(schemaLocation);
  }

  if (additionalSchemaLocations.length > 0) {
    args.push('--additional-schema-locations');
    args.push(additionalSchemaLocations.join(','));
  }

  if (!schemaLocation && additionalSchemaLocations.length === 0) {
    args.push('--schema-location');
    args.push('file://' + DEFAULT_SCHEMA_LOCATION);
  }

  if (ignoreMissingSchemas) {
    args.push('--ignore-missing-schemas');
  }

  if (skipKinds.length > 0) {
    args.push('--skip-kinds');
    args.push(skipKinds.join(','));
  }

  if (strict) {
    args.push('--strict');
  }
  return args;
}

function writeToStream(stream: Writable, data: string): Promise<void> {
  return new Promise((resolve, reject) =>
    stream.write(data, 'utf-8', (err) => (err ? reject(err) : resolve()))
  );
}

function readStdoutToString(childProcess: ChildProcess): Promise<string> {
  return new Promise<string>((resolve) => {
    let result = '';
    childProcess.stdout!!.on('data', (data) => {
      result += data.toString();
    });
    childProcess.on('close', () => {
      resolve(result);
    });
  });
}

kubeval.usage = `
Use kubeval to validate KRM resources against their json schemas.

The function configuration must be a ConfigMap.

The following keys can be used in the data field of the ConfigMap, and all of
them are optional:

${SCHEMA_LOCATION}: The base URL used to download the json schemas.
${ADDITIONAL_SCHEMA_LOCATIONS}: List of secondary base URLs used to download
  the json schemas.  These URLs will be used if the URL specified by
  ${SCHEMA_LOCATION} did not have the required schema.
${IGNORE_MISSING_SCHEMAS}: Skip validation for resources without a schema.  If
  omitted, a default value of false will be assumed.
${SKIP_KINDS}: Comma-separated list of case-sensitive kinds to skip when
  validating against schemas.  If omitted, no kinds will be skipped.
${STRICT}: Disallow additional properties that are not in the schemas.  If
  omitted, a default value of false will be assumed.
  
If neither ${SCHEMA_LOCATION} nor ${ADDITIONAL_SCHEMA_LOCATIONS} is provided, we
will convert the baked-in OpenAPI document to json schemas and use them.

Note: kpt fn render allow neither network access nor volume mount. That means
you need to use the baked-in OpenAPI schema when using this function in
kpt fn render.
`;

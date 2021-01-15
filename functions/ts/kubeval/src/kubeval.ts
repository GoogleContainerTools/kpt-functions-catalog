import {
  Configs,
  KubernetesObject,
  kubernetesObjectResult,
  Result,
} from 'kpt-functions';
import { ChildProcess, spawn } from 'child_process';
import { Writable } from 'stream';

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

  if (additionalSchemaLocations) {
    args.push('--additional-schema-locations');
    args.push(additionalSchemaLocations.join(','));
  }

  if (ignoreMissingSchemas) {
    args.push('--ignore-missing-schemas');
  }

  if (skipKinds) {
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
Validates configuration using kubeval.

Configured using a ConfigMap with the following keys, all of which are optional:
${SCHEMA_LOCATION}: The base URL used to download schemas.  If not specified,
  the default location at https://kubernetesjsonschema.dev/ will be used.
${ADDITIONAL_SCHEMA_LOCATIONS}: List of secondary base URLs used to download
  schemas.  These URLs will be used if the URL specified by ${SCHEMA_LOCATION}
  did not have the required schema.  By default, there are no secondary URLs,
  and only the primary base URL will be used.
${IGNORE_MISSING_SCHEMAS}: Skip validation for resource definitions without a
  schema.  If omitted, a default value of false will be assumed.
${SKIP_KINDS}: Comma-separated list of case-sensitive kinds to skip when
  validating against schemas.  If omitted, no kinds will be skipped.
${STRICT}: Disallow additional properties not in schema.  If omitted, a default
  value of false will be assumed.
`;

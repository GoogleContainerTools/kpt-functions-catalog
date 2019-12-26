/**
 * Copyright 2019 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import * as fs from 'fs';
import * as glob from 'glob';
import { safeLoadAll } from 'js-yaml';
import * as kpt from '@googlecontainertools/kpt-functions';
import * as path from 'path';

export const SOURCE_DIR = 'source_dir';
export const FILTER_IVNALID = 'filter_invalid';

export const readYaml: kpt.KptFunc = (configs) => {
  const sourceDir = configs.getFunctionConfigValueOrThrow(SOURCE_DIR);
  const ignoreInvalid = configs.getFunctionConfigValue(FILTER_IVNALID) === 'true';

  // Discard any input objects since this is a source function.
  configs.deleteAll();

  const files = glob.sync(sourceDir + '/**/*.+(yaml|yml)');
  const errors: kpt.ConfigError[] = files
    .map((f) => parseFile(configs, sourceDir, f, ignoreInvalid))
    .filter((err) => err !== undefined)
    .map((err) => err as kpt.ConfigError);

  if (errors.length) {
    return new kpt.MultiConfigError(
      `Found files containing invalid objects. To filter invalid objects set ${FILTER_IVNALID} to 'true'.`,
      errors,
    );
  }
  return;
};

readYaml.usage = `
Reads a directory of kubernetes YAML configs recursively.

Configured using a ConfigMap with the following keys:

${SOURCE_DIR}: Path to the config directory to read.
${FILTER_IVNALID}: [Optional] If 'true', ignores invalid Kubernetes objects instead of failing.

Example:

apiVersion: v1
kind: ConfigMap
data:
  ${SOURCE_DIR}: /path/to/source/dir
metadata:
  name: my-config
`;

function parseFile(
  configs: kpt.Configs,
  sourceDir: string,
  file: string,
  ignoreInvalid: boolean,
): kpt.ConfigError | undefined {
  const contents = readFileOrThrow(file);
  let objects = safeLoadAll(contents);

  const invalidObjects: object[] = objects.filter((o) => !kpt.isKubernetesObject(o));
  if (invalidObjects.length) {
    if (ignoreInvalid) {
      objects = objects.filter((o) => kpt.isKubernetesObject(o));
    } else {
      return new kpt.ConfigFileError(
        `File contains invalid Kubernetes objects '${JSON.stringify(invalidObjects)}'`,
        file,
      );
    }
  }

  objects.forEach((o, i) => {
    kpt.addAnnotation(o, kpt.SOURCE_PATH_ANNOTATION, path.relative(sourceDir, file));
    kpt.addAnnotation(o, kpt.SOURCE_INDEX_ANNOTATION, i.toString());
  });
  configs.insert(...objects);
  return;
}

function readFileOrThrow(f: string): string {
  try {
    return fs.readFileSync(f, 'utf8');
  } catch (err) {
    throw new Error(`Failed to read file ${f}: ${err}`);
  }
}

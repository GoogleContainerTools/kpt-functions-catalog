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
import * as kpt from 'kpt-functions';
import * as path from 'path';

export const SOURCE_DIR = new kpt.Param('source_dir', {
  help: 'Path to the source config directory',
  required: true,
});
export const FILTER_IVNALID = new kpt.Param('filter_invalid', {
  help: `If "true", Ignores objects that are not valid Kubernetes objects`,
  required: false,
});

/**
 * Reads a directory of kubernetes YAML configs recursively.
 *
 * An error is thrown if there are any exceptions reading files or parsing YAML.
 * Returns a ConfigError if any YAML file is not a KubernetesObject.
 */
export function readYAMLDir(configs: kpt.Configs) {
  const sourceDir = configs.getParam(SOURCE_DIR)!;
  const ignoreInvalid = configs.getParam(FILTER_IVNALID) === 'true';
  const files = glob.sync(sourceDir + '/**/*.+(yaml|yml)');

  // TODO(frankf): It's easy for source functions to not do this.
  // Explore way for framework to take care of this.
  configs.deleteAll();

  const errs: kpt.ConfigError[] = [];
  files.map((f) => {
    const err = parseFile(configs, sourceDir, f, ignoreInvalid);
    if (err) {
      errs.push(err);
    }
  });

  // TODO(willbeason): Provide way to return multiple errors since we want one error per file.
  return errs && errs[0];
}

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
      return new kpt.ConfigError(
        `File contains invalid Kubernetes objects ${file}: ${JSON.stringify(invalidObjects)}

To filter invalid objects using --filter_invalid flag.
        `,
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

export const RUNNER = kpt.Runner.newSource(readYAMLDir, SOURCE_DIR, FILTER_IVNALID);

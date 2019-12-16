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
import { existsSync, mkdirSync } from 'fs';
import * as glob from 'glob';
import { DumpOptions, safeDump } from 'js-yaml';
import * as kpt from 'kpt-functions';
import * as path from 'path';
import { isNamespace, Namespace } from './gen/io.k8s.api.core.v1';

export const SINK_DIR = new kpt.Param('sink_dir', {
  help: 'Path to the config directory to write to.',
  required: true,
});
export const OVERWRITE = new kpt.Param('overwrite', {
  help: `If "true", overwrite existing YAML files in sink_dir.
Otherwise, fail if any YAML files exist.`,
});

const DEFAULT_NAMESPACE = 'default';
const YAML_STYLE: DumpOptions = {
  /** indentation width to use (in spaces). */
  indent: 2,
  /** when true, will not add an indentation level to array elements */
  noArrayIndent: true,
};

/**
 * Write a directory of kubernetes YAML configs.
 *
 * If an object has the SOURCE_PATH_ANNOTATION, the file is created at that path.
 * Otherwise, the file is created at the top level dir using this convention for
 * the file name:
 *
 * |<namespace>/|<kind>_<name>.yaml
 *
 * If two objects have the same SOURCE_PATH_ANNOTATION, a multi-document file is
 * created. Ordering within this file is based on the order of objects in the input.
 */
export function writeYAMLDir(configs: kpt.Configs) {
  const sinkDir = configs.getParam(SINK_DIR)!;
  const overwrite = configs.getParam(OVERWRITE) === 'true';

  // Potential filesystem race conditions if another process is manipulating ths directory.
  const yamls = listYamlFiles(sinkDir);
  if (!overwrite && yamls.length > 0) {
    throw new Error(`--sink_dir contains YAML files and --overwrite is not set to "true".`);
  }

  const filesToDelete = new Set(yamls);

  configs.groupBy(buildSourcePath).forEach(([p, configsAtPath]) => {
    const documents = configsAtPath
      .sort(compareSourceIndex)
      .map((config) => kpt.removeAnnotation(config, kpt.SOURCE_INDEX_ANNOTATION))
      .map((config) => kpt.removeAnnotation(config, kpt.SOURCE_PATH_ANNOTATION))
      .map(toYaml);

    const file = path.join(sinkDir, p);
    const dir = path.dirname(file);
    if (!fs.existsSync(dir)) {
      fs.mkdirSync(path.dirname(file), { recursive: true });
    }
    const contents = documents.join('---\n');

    if (fs.existsSync(file)) {
      filesToDelete.delete(file);
      // Doesn't handle large files well. Should compare buffered output.
      const currentContents = fs.readFileSync(file).toString();
      if (contents == currentContents) {
        // No changes to make.
        return;
      }
    }

    fs.writeFileSync(file, contents, 'utf8');
  });

  filesToDelete.forEach((file) => {
    fs.unlinkSync(file);
  });
}

function listYamlFiles(dir: string): string[] {
  if (!existsSync(dir)) {
    mkdirSync(dir, { recursive: true });
  }
  // TODO(b/142947511): Does this work on Windows?
  return glob.sync(dir + '/**/*.+(yaml|yml)');
}

export const RUNNER = kpt.Runner.newSink(writeYAMLDir, SINK_DIR, OVERWRITE);

function toYaml(o: kpt.KubernetesObject): string {
  try {
    return safeDump(o, YAML_STYLE);
  } catch (err) {
    throw new Error(`Failed to write YAML for object: ${JSON.stringify(o)}: ${err}`);
  }
}

/**
 * Builds the fle path for the given object.
 *
 * If an object doesn't have the 'path' annotation, uses the convention:
 *
 * <namespace>/<name>_<kind>.yaml
 *
 * @param o The KubernetesObject to get a source path for.
 * @returns either the annotated source path, or a generated path for the object to be written to.
 */
// TODO(b/143073821): Solve general problem of making testing private methods unnecessary.
export function buildSourcePath(o: kpt.KubernetesObject): string {
  const annotationPath = kpt.getAnnotation(o, kpt.SOURCE_PATH_ANNOTATION);
  if (annotationPath) {
    return annotationPath;
  }

  if (isNamespace(o)) {
    // Special case to put Namespace objects in the same directory as the objects in them.
    return path.join(o.metadata.name, `${Namespace.kind.toLowerCase()}.yaml`);
  }

  let basePath = `${o.kind.toLowerCase()}_${o.metadata.name}.yaml`;
  if (o.metadata.namespace !== undefined) {
    // Namespace isn't undefined, so assume this is a Namespaced object. We don't yet support
    // distinguishing Namespaced and non-Namespaced resources any other way, and swagger.json does
    // not expose this information.
    let dir = o.metadata.namespace;
    if (dir === '') {
      // Namespace is explicitly set to empty string, so assume default Namespace like kubectl does.
      dir = DEFAULT_NAMESPACE;
    }
    return path.join(dir, basePath);
  }
  return basePath;
}

/**
 * Sorts the array of objects using 'index' annotation.
 *
 * If an object is missing index annotation, default to using zero.
 */
function compareSourceIndex(o1: kpt.KubernetesObject, o2: kpt.KubernetesObject): number {
  const i1 = Number(kpt.getAnnotation(o1, kpt.SOURCE_INDEX_ANNOTATION)) | 0;
  const i2 = Number(kpt.getAnnotation(o2, kpt.SOURCE_INDEX_ANNOTATION)) | 0;
  return i1 - i2;
}

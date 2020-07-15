/**
 * Copyright 2020 Google LLC
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

import { Configs, isKubernetesObject, generalResult } from 'kpt-functions';
import { spawnSync } from 'child_process';
import { safeLoadAll } from 'js-yaml';

const BUILD_PATH = 'path';

export async function kustomizeBuild(configs: Configs) {
  // Validate config data and read arguments.
  const args = readArguments(configs);
  args.unshift('build');

  let error;
  try {
    const child = spawnSync('kustomize', args);
    error = child.stderr;
    let objects = safeLoadAll(child.stdout);
    objects = objects.filter((o) => isKubernetesObject(o));
    configs.insert(...objects);
  } catch (err) {
    configs.addResults(generalResult(err, 'error'));
  }
  if (error && error.length > 0) {
    configs.addResults(
      generalResult(
        `Kustomize build command results in error: ${error.toString()}`,
        'error'
      )
    );
  }
}

function readArguments(configs: Configs) {
  const args: string[] = [];
  const configMap = configs.getFunctionConfigMap();
  if (!configMap) {
    return args;
  }
  configMap.forEach((value: string, key: string) => {
    if (key === BUILD_PATH) {
      args.push(value);
    } else {
      args.push(key);
      args.push(value);
    }
  });
  return args;
}

kustomizeBuild.usage = `
Build Kubernetes manifests using kustomize build. 

Configured using a ConfigMap with a key for {${BUILD_PATH}}.
Works with arbitrary kustomize build flags like --reorder:

${BUILD_PATH}: [Optional, default '.'] Path to kustomization.yaml.
--reorder: [Optional] Reorder the resources just before output.
...

Example:

To build a kustomization at '/path/to/kustomization' using '--reorder none':

apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  annotations:
    config.k8s.io/function: |
      container:
        image:  gcr.io/kpt-functions/kustomize-build
    config.kubernetes.io/local-config: "true"
data:
  ${BUILD_PATH}: /path/to/kustomization
  --reorder: none
`;

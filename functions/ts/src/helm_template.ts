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

import { safeLoadAll } from 'js-yaml';
import {
  Configs,
  FunctionConfigError,
  isKubernetesObject,
  generalResult,
} from 'kpt-functions';
import { spawnSync } from 'child_process';
import { isConfigMap } from './gen/io.k8s.api.core.v1';

const CHART_NAME = 'name';
const CHART_PATH = 'chart_path';
const VALUES_PATH = '--values';

// Render chart templates locally using helm template.
export async function helmTemplate(configs: Configs) {
  // Validate config data and read arguments.
  const args = readArguments(configs);
  args.unshift('template');

  let error;
  try {
    const child = spawnSync('helm', args);
    error = child.stderr;
    let objects = safeLoadAll(child.stdout);
    objects = objects.filter(o => isKubernetesObject(o));
    configs.insert(...objects);
  } catch (err) {
    configs.addResults(generalResult(err, 'error'));
  }
  if (error && error.length > 0) {
    configs.addResults(
      generalResult(`Helm template command results in error: ${error}`, 'error')
    );
  }
}

function readArguments(configs: Configs) {
  const args: string[] = [];
  let nameArg;
  let pathArg;
  const data = readConfigDataOrThrow(configs);
  for (const key in data) {
    if (key === CHART_NAME) {
      nameArg = data[key];
    } else if (key === CHART_PATH) {
      pathArg = data[key];
    } else {
      args.push(key);
      args.push(data[key]);
    }
  }

  // Helm template expects name and chart path first so place those at the beginning
  if (pathArg) {
    args.unshift(pathArg);
  }
  if (nameArg) {
    args.unshift(nameArg);
  }

  return args;
}

function readConfigDataOrThrow(configs: Configs) {
  const cm = configs.getFunctionConfig();
  if (!cm) {
    throw new FunctionConfigError(`functionConfig expected, instead undefined`);
  }
  if (!isConfigMap(cm)) {
    throw new FunctionConfigError(
      `functionConfig expected to be of kind ConfigMap, instead got: ${cm.kind}`
    );
  }
  if (!cm.data) {
    throw new FunctionConfigError(
      `functionConfig expected to contain data, instead empty`
    );
  }
  return cm.data;
}

helmTemplate.usage = `
Render chart templates locally using helm template. If input a list of configs in
addition to arguments will overwrite any chart objects that already exist in the list.

Configured using a ConfigMap with keys for ${CHART_NAME}, ${CHART_PATH}.
Works with arbitrary helm template flags like --values:

${CHART_NAME}: Name of helm chart.
${CHART_PATH}: Chart templates directory.
${VALUES_PATH}: [Optional] Path to values file.
...

Example:

To expand a chart named 'my-chart' at '../path/to/helm/chart' using './values.yaml':

apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  annotations:
    config.k8s.io/function: |
      container:
        image:  gcr.io/kpt-functions/helm-template
    config.kubernetes.io/local-config: "true"
data:
  ${CHART_NAME}: my-chart
  ${CHART_PATH}: ../path/to/helm/chart
  ${VALUES_PATH}: ./values.yaml
`;

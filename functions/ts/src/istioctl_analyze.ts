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

import {
  Configs,
  FunctionConfigError,
  configFileResult,
  Severity,
} from 'kpt-functions';
import { spawnSync } from 'child_process';
import { isConfigMap } from './gen/io.k8s.api.core.v1';

const FILE_ARGS = 'files';
const FLAG_ARGS = 'flags';
const USE_KUBE_FLAG = '--use-kube';
const OUTPUT_SHORT_FLAG = '-o';
const OUTPUT_LONG_FLAG = '--output';

interface IstioResult {
  code: string;
  level: 'Error' | 'Warn' | 'Info';
  origin: string;
  reference: string;
  message: string;
  documentation_url: string;
}

// Analyze istio configs using istioctl analyze.
export async function istioctlAnalyze(configs: Configs) {
  // Validate config data and read arguments.
  const args = readArguments(configs);
  args.unshift('analyze');

  let error;
  try {
    const child = spawnSync('istioctl', args);
    error = child.stderr;
    if (child.stdout && child.stdout !== 'null') {
      const istioOutput: IstioResult[] = JSON.parse(child.stdout);
      if (istioOutput && istioOutput.length) {
        istioOutput.forEach(istioResult => {
          const result = configFileResult(
            istioResult.message,
            istioResult.reference,
            istioResult.level.toLowerCase() as Severity
          );
          result.tags = {
            ['documentation_url']: istioResult.documentation_url,
            ['origin']: istioResult.origin,
            ['code']: istioResult.code,
          };
          configs.addResults(result);
        });
      }
    }
  } catch (err) {
    configs.addResults(configFileResult(`${err}`, '', 'error'));
  }
  if (error && error.length > 0) {
    configs.addResults(
      configFileResult(
        `Istioctl analyze command results in error: ${error}`,
        '',
        'error'
      )
    );
  }
}

function readArguments(configs: Configs) {
  // Initialize to output json
  const args: string[] = ['-o', 'json'];
  const data = readConfigDataOrThrow(configs);
  for (const key in data) {
    if (key === FILE_ARGS || key === FLAG_ARGS) {
      args.push(data[key]);
    } else if (key === OUTPUT_SHORT_FLAG || key === OUTPUT_LONG_FLAG) {
      continue;
    } else if (key === USE_KUBE_FLAG) {
      // use-kube flag which needs equals sign instead of space separator
      args.push(`${key}=${data[key]}`);
    } else if (data.hasOwnProperty(key)) {
      args.push(key);
      args.push(data[key]);
    }
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

istioctlAnalyze.usage = `
Istioctl analyze is a diagnostic tool that can detect potential issues with your
Istio configuration and output errors to the results field. This function runs
against local configuration files to catch problems before you apply changes to a
cluster.

Configure this function using a ConfigMap with keys for "${FILE_ARGS}", "${FLAG_ARGS}", and
arbitrary istioctl analyze flags. The "${FILE_ARGS}" argument takes an array of
files and directories to analyze. The "${FLAG_ARGS}" argument takes an array of
flags which do not take arguments. Arbitrary istioctl analyze flags which take their
own arguments, like --suppress, should be passed as separate arguments. The --output
flag is ignored as all output is included in config results.

Accepted arguments:
${FILE_ARGS}: [Required] List of file or directory arguments to istioctl analyze.
${FLAG_ARGS}: [Optional] List of flag arguments to istioctl analyze.
...

Example: Analyze '/path/to/istio/configs' recursively using '--use-kube=false'
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  annotations:
    config.k8s.io/function: |
      container:
        image:  gcr.io/kpt-functions/istioctl-analyze
    config.kubernetes.io/local-config: "true"
data:
  "${FILE_ARGS}": ["/path/to/istio/configs"]
  "${FLAG_ARGS}": ["--recursive"]
  "${USE_KUBE_FLAG}": "false"
`;

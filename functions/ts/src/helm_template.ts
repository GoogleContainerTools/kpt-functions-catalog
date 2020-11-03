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
import { Configs, isKubernetesObject, generalResult } from 'kpt-functions';
import { spawnSync } from 'child_process';
import { tmpdir } from 'os';
import Path from 'path';
import { mkdirSync, rmdirSync, readdirSync } from 'fs';

const CHART_NAME = 'name';
const LOCAL_CHART_PATH = 'local-chart-path';
const VALUES_PATH = '--values';
// CHART is chart name that will be processed.
const CHART = 'chart';
// CHART_REPO is the repo name that charts are pulled from
const CHART_REPO = 'chart-repo';
// CHART_REPO_URL is the repo remote URL
const CHART_REPO_URL = 'chart-repo-url';

class HelmError extends Error {
  constructor(m: string) {
    super(m);

    // Set the prototype explicitly.
    Object.setPrototypeOf(this, HelmError.prototype);
  }
}

// get and validate the data field from function ConfigMap
function getConfigMapData(configs: Configs): Map<string, string> {
  const configMapData = configs.getFunctionConfigMap();
  if (configMapData === undefined) {
    throw new Error(`Function ConfigMap data cannot be undefined.`);
  }
  validateConfigMapData(configMapData);
  return configMapData;
}

function validateConfigMapData(configMapData: Map<string, string>) {
  // either local or remote
  if (!configMapData.has(CHART) && !configMapData.has(LOCAL_CHART_PATH)) {
    throw new Error(
      `Either ${CHART} or ${LOCAL_CHART_PATH} needs to be provided.`
    );
  }
  // cannot use both remote and local chart
  if (configMapData.has(CHART) && configMapData.has(LOCAL_CHART_PATH)) {
    throw new Error(
      `Cannot use ${CHART} and ${LOCAL_CHART_PATH} at the same time.`
    );
  }
  // CHART_REPO and CHART_REPO_URL are required for remote chart
  if (
    configMapData.has(CHART) &&
    (!configMapData.has(CHART_REPO) || !configMapData.has(CHART_REPO_URL))
  ) {
    throw new Error(
      `${CHART_REPO} and ${CHART_REPO_URL} are required for remote chart`
    );
  }
}

// run helm repo add to add the remote repo url to helm repo list
async function addHelmRepo(configMapData: Map<string, string>) {
  const args: string[] = [
    'repo',
    'add',
    configMapData.get(CHART_REPO)!,
    configMapData.get(CHART_REPO_URL)!,
  ];
  runHelmCommand(args);
}

// mkTmpDir creates a temporary directory and return the path
async function mkTmpDir(): Promise<string> {
  const path = Path.join(tmpdir(), 'remote-helm-chart');
  // clear at first
  rmdirSync(path, { recursive: true });

  mkdirSync(path, { recursive: true });
  return path;
}

// run the helm command with given args
function runHelmCommand(args: string[]): { stdout: string; stderr: string } {
  const child = spawnSync('helm', args);
  const stderr = child.stderr;
  const error = child.error;

  if (error || (stderr && stderr.length > 0)) {
    throw new HelmError(
      `Helm command ${args.join(' ')} results in error: ${stderr.toString()}`
    );
  }

  return {
    stdout: child.stdout,
    stderr,
  };
}

// run helm pull command
async function runHelmPull(configMapData: Map<string, string>) {
  if (!configMapData.has(CHART)) {
    // not remote chart
    return;
  }
  addHelmRepo(configMapData);
  // prepare source directory
  const tmpDir = await mkTmpDir();

  const args: string[] = [
    'pull',
    '--untar',
    '--untardir',
    tmpDir,
    configMapData.get(CHART)!,
  ];

  runHelmCommand(args);

  // helm pull will untar the charts to a subdirectory in destination
  // we need to use that subdirectory as LOCAL_CHART_PATH
  const [subDir] = readdirSync(tmpDir);
  configMapData.set(LOCAL_CHART_PATH, Path.join(tmpDir, subDir));
}

// get arguments for helm template command
function getTemplateArgs(configMapData: Map<string, string>): string[] {
  const args: string[] = [];
  configMapData.forEach((value: string, key: string) => {
    // template flags should start with '-'
    if (!key.startsWith('-')) {
      return;
    }
    args.push(key);
    args.push(value);
  });

  // Helm template expects name and chart path first so place those at the beginning
  if (configMapData.get(LOCAL_CHART_PATH) !== undefined) {
    args.unshift(configMapData.get(LOCAL_CHART_PATH)!);
  }
  if (configMapData.get(CHART_NAME) !== undefined) {
    args.unshift(configMapData.get(CHART_NAME)!);
  }

  return args;
}

// run helm template command
async function runHelmTemplate(
  configs: Configs,
  configMapData: Map<string, string>
) {
  // Validate config data and read arguments.
  const args = getTemplateArgs(configMapData);
  args.unshift('template');

  const { stdout } = runHelmCommand(args);

  try {
    let objects = safeLoadAll(stdout);
    objects = objects.filter((o) => isKubernetesObject(o));
    configs.insert(...objects);
  } catch (err) {
    throw new HelmError(err);
  }
}

// Render local ot remote chart templates using helm template.
export async function helmTemplate(configs: Configs) {
  try {
    const configMapData = getConfigMapData(configs);
    await runHelmPull(configMapData);
    await runHelmTemplate(configs, configMapData);
  } catch (err) {
    if (err instanceof HelmError) {
      configs.addResults(generalResult(err.toString(), 'error'));
    } else {
      throw err;
    }
  }
}

helmTemplate.usage = `
Render chart templates locally using local or remote helm template. If input a list of configs in
addition to arguments will overwrite any chart objects that already exist in the list.

If ${CHART} is given, the function will add the ${CHART_REPO} with ${CHART_REPO_URL} to helm repo list. 
Then pull the ${CHART} to a local directory. If a ${CHART} is not given, the function expects an unpacked chart 
in the directory ${LOCAL_CHART_PATH}.

Configured using a ConfigMap with keys for ${CHART_NAME}, ${LOCAL_CHART_PATH}, ${CHART}, ${CHART_REPO} or ${CHART_REPO_URL}.
Works with arbitrary helm template flags like --values:

${CHART_NAME}: [Optional] Name of helm chart.
${LOCAL_CHART_PATH}: [Optional] Chart templates directory.
${VALUES_PATH}: [Optional] Path to values file.
${CHART}: [Optional] A url to pull the remote chart instead of using local templates. This flag only works with remote charts.
${CHART_REPO}: [Optional] Repo name that helm should pull the templates from. Only used when chart is provided.
${CHART_REPO_URL}: [Optional] Repo list URL which will be added to the helm repo list with repo name chart-repo. Only used when chart is provided.
...

Examples:

1. To expand a chart named 'my-chart' at remote chart 'stable/chart'

apiVersion: v1
kind: ConfigMap
metadata:
  name: my-func-config
  annotations:
    config.kubernetes.io/function: |
      container:
        image: gcr.io/kpt-functions/helm-template
        network:
          required: true
    config.kubernetes.io/local-config: "true"
data:
  ${CHART_REPO}: stable
  ${CHART_REPO_URL}: https://url/to/repo
  ${CHART}: stable/chart
  ${CHART_NAME}: my-chart

2. To expand a chart named 'my-chart' at '../path/to/helm/chart' using './values.yaml':

apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  annotations:
    config.k8s.io/function: |
      container:
        image: gcr.io/kpt-functions/helm-template
    config.kubernetes.io/local-config: "true"
data:
  ${CHART_NAME}: my-chart
  ${LOCAL_CHART_PATH}: ../path/to/helm/chart
  ${VALUES_PATH}: ./values.yaml
`;

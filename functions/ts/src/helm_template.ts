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

class HelmTemplateError extends Error {
  constructor(m: string) {
    super(m);

    // Set the prototype explicitly.
    Object.setPrototypeOf(this, HelmTemplateError.prototype);
  }
}

class HelmTemplate {
  configs: Configs;
  data: Map<string, string> = new Map<string, string>();

  constructor(configs: Configs) {
    this.configs = configs;

    this.getData();
  }

  getData() {
    const data = this.configs.getFunctionConfigMap();
    if (data !== undefined) {
      this.data = data;
    }
    this.validData();
  }

  validData() {
    // either local or remote
    if (!this.data.has(CHART) && !this.data.has(LOCAL_CHART_PATH)) {
      throw new Error(
        `Either ${CHART} or ${LOCAL_CHART_PATH} needs to be provided.`
      );
    }
    // cannot use both remote and local chart
    if (this.data.has(CHART) && this.data.has(LOCAL_CHART_PATH)) {
      throw new Error(
        `Cannot use ${CHART} and ${LOCAL_CHART_PATH} at the same time.`
      );
    }
    // CHART_REPO and CHART_REPO_URL are required for remote chart
    if (
      this.data.has(CHART) &&
      (!this.data.has(CHART_REPO) || !this.data.has(CHART_REPO_URL))
    ) {
      throw new Error(
        `${CHART_REPO} and ${CHART_REPO_URL} are required for remote chart`
      );
    }
  }

  // get arguments for helm template command
  templateArgs(): string[] {
    const args: string[] = [];
    this.data.forEach((value: string, key: string) => {
      // template flags should start with '-'
      if (!key.startsWith('-')) {
        return;
      }
      args.push(key);
      args.push(value);
    });

    // Helm template expects name and chart path first so place those at the beginning
    if (this.data.get(LOCAL_CHART_PATH) !== undefined) {
      args.unshift(this.data.get(LOCAL_CHART_PATH)!);
    }
    if (this.data.get(CHART_NAME) !== undefined) {
      args.unshift(this.data.get(CHART_NAME)!);
    }

    return args;
  }

  // run helm template command
  async template() {
    // Validate config data and read arguments.
    const args = this.templateArgs();
    args.unshift('template');

    const { stdout } = this.runHelmCommand(args);

    try {
      let objects = safeLoadAll(stdout);
      objects = objects.filter((o) => isKubernetesObject(o));
      this.configs.insert(...objects);
    } catch (err) {
      throw new HelmTemplateError(err);
    }
  }

  // run helm pull command
  async pull() {
    if (!this.data.has(CHART)) {
      // not remote chart
      return;
    }
    this.addRepo();
    // prepare source directory
    const tmpDir = await this.mkTmpDir();

    const args: string[] = [
      'pull',
      '--untar',
      '--untardir',
      tmpDir,
      this.data.get(CHART)!,
    ];

    this.runHelmCommand(args);

    // helm pull will untar the charts to a subdirectory in destination
    // we need to use that subdirectory as LOCAL_CHART_PATH
    const [subDir] = readdirSync(tmpDir);
    this.data.set(LOCAL_CHART_PATH, Path.join(tmpDir, subDir));
  }

  // run helm repo add to add the remote repo url to helm repo list
  async addRepo() {
    const args: string[] = [
      'repo',
      'add',
      this.data.get(CHART_REPO)!,
      this.data.get(CHART_REPO_URL)!,
    ];
    this.runHelmCommand(args);
  }

  // mkTmpDir creates a temporary directory and return the path
  async mkTmpDir(): Promise<string> {
    const path = Path.join(tmpdir(), 'remote-helm-chart');
    // clear at first
    rmdirSync(path, { recursive: true });

    mkdirSync(path, { recursive: true });
    return path;
  }

  // run the helm command with given args
  runHelmCommand(args: string[]): { stdout: string; stderr: string } {
    const child = spawnSync('helm', args);
    const stderr = child.stderr;
    const error = child.error;

    if (error || (stderr && stderr.length > 0)) {
      throw new HelmTemplateError(
        `Helm command ${args.join(' ')} results in error: ${stderr.toString()}`
      );
    }

    return {
      stdout: child.stdout,
      stderr,
    };
  }
}

// Render local ot remote chart templates using helm template.
export async function helmTemplate(configs: Configs) {
  try {
    const helm = new HelmTemplate(configs);
    await helm.pull();
    await helm.template();
  } catch (err) {
    if (err instanceof HelmTemplateError) {
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

${CHART_NAME}: Name of helm chart.
${LOCAL_CHART_PATH}: Chart templates directory.
${VALUES_PATH}: [Optional] Path to values file.
${CHART}: A url to pull the remote chart instead of using local templates. This flag only works with remote charts.
${CHART_REPO}: Repo name that helm should pull the templates from. Only used when chart is provided.
${CHART_REPO_URL}: Repo list URL which will be added to the helm repo list with repo name chart-repo. Only used when chart is provided.
...

Examples:

1. To expand a chart named 'my-chart' at '../path/to/helm/chart' using './values.yaml':

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

2. To expand a chart named 'my-chart' at remote chart 'stable/chart'

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
  chart-repo: stable
  chart-repo-url: https://url/to/repo
  chart: stable/chart
`;

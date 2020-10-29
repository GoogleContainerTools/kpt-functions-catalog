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

import { Configs, TestRunner, KptFunc, generalResult } from 'kpt-functions';
import { helmTemplate } from './helm_template';
import { Namespace, ConfigMap } from './gen/io.k8s.api.core.v1';
import * as proxyquire from 'proxyquire';

describe('helmTemplate', () => {
  const runner = new TestRunner(helmTemplate);
  it('outputs error given namespace function config', async () => {
    const namespace = Namespace.named('namespace');
    const input = new Configs(undefined, namespace);

    await runner.assert(input, new Configs(undefined, namespace), Error);
  });
});

describe('helm tamplate happy path', () => {
  let helmTemplateMocked: { helmTemplate: KptFunc };
  let runner: TestRunner;
  beforeAll(() => {
    helmTemplateMocked = proxyquire.noCallThru()('./helm_template', {
      child_process: {
        spawnSync: (command: string, args: string[]) => {
          switch (args[0]) {
            case 'template':
              return {
                stdout: `
apiVersion: v1
kind: Namespace
metadata:
  name: ns
---
foo: bar
`,
                stderr: '',
                error: undefined,
              };
            case 'pull':
              return {
                stdout: '',
                stderr: '',
                error: undefined,
              };
            case 'repo':
              return {
                stdout: '',
                stderr: '',
                error: undefined,
              };
            default:
              return {
                stdout: '',
                stderr: 'Unkown command',
                error: undefined,
              };
          }
        },
      },
      fs: {
        mkdirSync: () => {},
        rmdirSync: () => {},
        readdirSync: () => ['tmpdir'],
      },
    });
    runner = new TestRunner(helmTemplateMocked.helmTemplate);
  });

  it('helm template local happy', async () => {
    const configMap = ConfigMap.named('config');
    configMap.data = {
      'local-chart-path': 'bar',
      '--values': 'whatever',
      name: 'chart-name',
    };
    const input = new Configs(undefined, configMap);
    const output = input.deepCopy();
    output.insert(Namespace.named('ns'));
    await runner.assert(input, output);
  });

  it('helm template remote happy', async () => {
    const configMap = ConfigMap.named('config');
    configMap.data = {
      chart: 'stable/chart',
      'chart-repo-url': 'https://url/to/repo',
      'chart-repo': 'stable',
    };
    const input = new Configs(undefined, configMap);
    const output = input.deepCopy();
    output.insert(Namespace.named('ns'));
    await runner.assert(input, output);
  });
});

describe('invalid config', () => {
  it('not local or remote', async () => {
    const runner = new TestRunner(helmTemplate);
    const configMap = ConfigMap.named('config');
    const input = new Configs(undefined, configMap);
    const output = input.deepCopy();
    await runner.assert(
      input,
      output,
      Error,
      /Either .*? or .*? needs to be provided/
    );
  });

  it('no function config', async () => {
    const runner = new TestRunner(helmTemplate);
    const input = new Configs(undefined, undefined);
    const output = input.deepCopy();
    await runner.assert(
      input,
      output,
      Error,
      'Function ConfigMap data cannot be undefined.'
    );
  });

  it('local and remote', async () => {
    const runner = new TestRunner(helmTemplate);
    const configMap = ConfigMap.named('config');
    configMap.data = {
      chart: 'stable/chart',
      'local-chart-path': 'foo',
    };
    const input = new Configs(undefined, configMap);
    const output = input.deepCopy();
    await runner.assert(
      input,
      output,
      Error,
      /Cannot use .*? and .*? at the same time/
    );
  });

  it('lack remote info', async () => {
    const runner = new TestRunner(helmTemplate);
    const configMap = ConfigMap.named('config');
    configMap.data = {
      chart: 'stable/chart',
    };
    const input = new Configs(undefined, configMap);
    const output = input.deepCopy();
    await runner.assert(
      input,
      output,
      Error,
      /.*? and .*? are required for remote chart/
    );
  });

  it('lack remote repo url', async () => {
    const runner = new TestRunner(helmTemplate);
    const configMap = ConfigMap.named('config');
    configMap.data = {
      chart: 'stable/chart',
      'chart-repo': 'stable',
    };
    const input = new Configs(undefined, configMap);
    const output = input.deepCopy();
    await runner.assert(
      input,
      output,
      Error,
      /.*? and .*? are required for remote chart/
    );
  });
});

describe('error when run helm command', () => {
  let helmTemplateMocked: { helmTemplate: KptFunc };
  let runner: TestRunner;
  beforeAll(() => {
    helmTemplateMocked = proxyquire.noCallThru()('./helm_template', {
      child_process: {
        spawnSync: (command: string, args: string[]) => {
          return {
            stdout: '',
            stderr: 'helm error',
            error: undefined,
          };
        },
      },
    });
    runner = new TestRunner(helmTemplateMocked.helmTemplate);
  });

  it('error when run helm command', async () => {
    const configMap = ConfigMap.named('config');
    configMap.data = {
      'local-chart-path': 'bar',
      '--values': 'whatever',
      name: 'chart-name',
    };
    const input = new Configs(undefined, configMap);
    const output = input.deepCopy();
    output.addResults(
      generalResult(
        'Error: Helm command template chart-name bar --values whatever results in error: helm error',
        'error'
      )
    );
    await runner.assert(input, output);
  });
});

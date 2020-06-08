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

import { Configs, TestRunner, FunctionConfigError } from 'kpt-functions';
import { istioctlAnalyze } from './istioctl_analyze';
import { Namespace, ConfigMap } from './gen/io.k8s.api.core.v1';

const RUNNER = new TestRunner(istioctlAnalyze);

describe('istioctlAnalyze', () => {
  it('outputs error given undefined function config', async () => {
    const input = new Configs(undefined, undefined);

    await RUNNER.assert(
      input,
      new Configs(undefined),
      FunctionConfigError,
      'functionConfig expected, instead undefined'
    );
  });

  const namespace = Namespace.named('namespace');
  it('outputs error given namespace function config', async () => {
    const input = new Configs(undefined, namespace);

    await RUNNER.assert(
      input,
      new Configs(undefined, namespace),
      FunctionConfigError,
      'functionConfig expected to be of kind ConfigMap, instead got: Namespace'
    );
  });

  const emptyConfigMap = new ConfigMap({ metadata: { name: 'config' } });
  it('outputs error given empty function config', async () => {
    const input = new Configs(undefined, emptyConfigMap);

    await RUNNER.assert(
      input,
      new Configs(undefined, emptyConfigMap),
      FunctionConfigError,
      'functionConfig expected to contain data, instead empty'
    );
  });

  const configMap = new ConfigMap({
    metadata: { name: 'config' },
    data: { '--use-kube': 'false' },
  });
  it('keeps configs unchanged', async () => {
    const input = new Configs(undefined, configMap);
    const output = new Configs(undefined, configMap);
    await RUNNER.assert(input, output);
  });
});

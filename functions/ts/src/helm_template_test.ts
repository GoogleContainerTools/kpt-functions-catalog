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
import { helmTemplate } from './helm_template';
import { Namespace } from './gen/io.k8s.api.core.v1';

const RUNNER = new TestRunner(helmTemplate);

describe('helmTemplate', () => {
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

    await RUNNER.assert(input, new Configs(undefined, namespace), Error);
  });
});

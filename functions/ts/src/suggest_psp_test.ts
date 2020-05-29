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

import { Configs, TestRunner, kubernetesObjectResult } from 'kpt-functions';
import { PodSecurityPolicy } from './gen/io.k8s.api.policy.v1beta1';
import { suggestPsp } from './suggest_psp';

const RUNNER = new TestRunner(suggestPsp);

describe('suggestPsp', () => {
  it('empty configs is noop', RUNNER.assertCallback(undefined, 'unchanged'));

  it(
    'suggest PSP with allowPrivilegeEscalation = true to false',
    RUNNER.assertCallback(
      new Configs([psp(true)]),
      new Configs([psp(true)], undefined, [
        kubernetesObjectResult(
          'Suggest explicitly disabling privilege escalation',
          psp(true),
          { path: 'spec.allowPrivilegeEscalation', suggestedValue: false },
          'warn',
          { category: 'security' }
        ),
      ])
    )
  );

  it(
    'leaves PSP with allowPrivilegeEscalation = false alone',
    RUNNER.assertCallback(new Configs([psp(false)]), 'unchanged')
  );
});

function psp(allowPrivilegeEscalation: boolean): PodSecurityPolicy {
  return new PodSecurityPolicy({
    metadata: {
      name: 'pod',
    },
    spec: {
      allowPrivilegeEscalation,
      fsGroup: { rule: 'RunAsAny' },
      runAsUser: { rule: 'RunAsAny' },
      seLinux: { rule: 'RunAsAny' },
      supplementalGroups: { rule: 'RunAsAny' },
    },
  });
}

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

import { Configs, newManifestError, TestRunner } from '@googlecontainertools/kpt-functions';
import { ClusterRoleBinding, RoleBinding, Subject } from './gen/io.k8s.api.rbac.v1';
import { validateRolebinding, SUBJECT_NAME } from './validate_rolebinding';
import { ConfigMap } from './gen/io.k8s.api.core.v1';

function roleBinding(name: string, ...subjects: Subject[]): RoleBinding {
  return new RoleBinding({
    metadata: { name },
    roleRef: {
      apiGroup: 'rbac',
      kind: 'Role',
      name: 'alice',
    },
    subjects: subjects,
  });
}

const RUNNER = new TestRunner(validateRolebinding);

describe(validateRolebinding.name, () => {
  let functionConfig = ConfigMap.named('config');
  functionConfig.data = {};
  functionConfig.data![SUBJECT_NAME] = 'alice@example.com';

  it('passes empty input', RUNNER.run(undefined, undefined, true));

  it(
    'passes valid RoleBindings',
    RUNNER.run(
      new Configs(
        [
          roleBinding('alice', {
            name: 'backend-all@example.com',
            kind: 'User',
          }),
        ],
        functionConfig,
      ),
    ),
  );

  it(
    'fails invalid RoleBindings',
    RUNNER.run(
      new Configs(
        [
          roleBinding('alice', {
            name: 'alice@example.com',
            kind: 'User',
          }),
        ],
        functionConfig,
      ),
      newManifestError('Found RoleBindings with banned subjects'),
    ),
  );

  it(
    'ignores ClusterRoleBinding subjects',
    RUNNER.run(
      new Configs(
        [
          new ClusterRoleBinding({
            metadata: { name: 'alice' },
            roleRef: {
              apiGroup: 'rbac',
              kind: 'Role',
              name: 'alice',
            },
            subjects: [
              {
                name: 'alice@example.com',
                kind: 'User',
              },
            ],
          }),
        ],
        functionConfig,
      ),
    ),
  );
});

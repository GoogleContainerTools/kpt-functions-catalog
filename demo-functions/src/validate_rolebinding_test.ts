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

import { Configs, newManifestError, TestRunner } from 'kpt-functions';
import { ClusterRoleBinding, RoleBinding, Subject } from './gen/io.k8s.api.rbac.v1';
import { disallowRoleBindingSubject, SUBJECT_NAME } from './validate_rolebinding';

function roleBinding(name: string, ...subjects: Subject[]): RoleBinding {
  return new RoleBinding({
    metadata: { name },
    roleRef: {
      apiGroup: 'rbac',
      kind: 'Role',
      name: 'admin',
    },
    subjects: subjects,
  });
}

const RUNNER = new TestRunner(disallowRoleBindingSubject);

describe(disallowRoleBindingSubject.name, () => {
  it('passes empty input', RUNNER.run());

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
        new Map([[SUBJECT_NAME.name, 'admin@example.com']]),
      ),
    ),
  );

  it(
    'fails invalid RoleBindings',
    RUNNER.run(
      new Configs(
        [
          roleBinding('alice', {
            name: 'admin@example.com',
            kind: 'User',
          }),
        ],
        new Map([[SUBJECT_NAME.name, 'admin@example.com']]),
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
              name: 'admin',
            },
            subjects: [
              {
                name: 'admin@example.com',
                kind: 'User',
              },
            ],
          }),
        ],
        new Map([[SUBJECT_NAME.name, 'admin@example.com']]),
      ),
    ),
  );
});

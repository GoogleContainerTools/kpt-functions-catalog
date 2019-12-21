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

import { Configs, TestRunner } from '@googlecontainertools/kpt-functions';
import { Team } from './gen/dev.cft.anthos.v1alpha1';
import { Namespace } from './gen/io.k8s.api.core.v1';
import { ClusterRole, RoleBinding, Subject } from './gen/io.k8s.api.rbac.v1';
import { hydrateAnthosTeam } from './hydrate_anthos_team';

function team(name: string, ...roles: Team.Spec.Item[]): Team {
  const team = new Team({
    metadata: { name },
    spec: {},
  });
  if (roles.length) {
    team.spec.roles = roles;
  }
  return team;
}

const RUNNER = new TestRunner(hydrateAnthosTeam);

describe(hydrateAnthosTeam.name, () => {
  it('does nothing to empty repos', RUNNER.run());

  it('does nothing to non-Team objects', RUNNER.run(new Configs([Namespace.named('backend')])));

  it(
    'expands an empty team to its environments',
    RUNNER.run(
      new Configs([team('backend')]),
      new Configs([
        team('backend'),
        Namespace.named('backend-dev'),
        Namespace.named('backend-prod'),
      ]),
    ),
  );

  it(
    'expands a Team with group roles',
    RUNNER.run(
      new Configs([
        team('backend', {
          role: 'admin',
          groups: ['sre@example.com'],
        }),
      ]),
      new Configs([
        team('backend', {
          role: 'admin',
          groups: ['sre@example.com'],
        }),
        Namespace.named('backend-dev'),
        new RoleBinding({
          metadata: {
            name: 'admin',
            namespace: 'backend-dev',
          },
          roleRef: {
            kind: ClusterRole.kind,
            apiGroup: 'rbac.authorization.k8s.io',
            name: 'admin',
          },
          subjects: [
            new Subject({
              kind: 'Group',
              name: 'sre@example.com',
            }),
          ],
        }),
        Namespace.named('backend-prod'),
        new RoleBinding({
          metadata: {
            name: 'admin',
            namespace: 'backend-prod',
          },
          roleRef: {
            kind: ClusterRole.kind,
            apiGroup: 'rbac.authorization.k8s.io',
            name: 'admin',
          },
          subjects: [
            new Subject({
              kind: 'Group',
              name: 'sre@example.com',
            }),
          ],
        }),
      ]),
    ),
  );

  it(
    'expands a Team with user roles',
    RUNNER.run(
      new Configs([
        team('backend', {
          role: 'admin',
          users: ['admin@example.com'],
        }),
      ]),
      new Configs([
        team('backend', {
          role: 'admin',
          users: ['admin@example.com'],
        }),
        Namespace.named('backend-dev'),
        new RoleBinding({
          metadata: {
            name: 'admin',
            namespace: 'backend-dev',
          },
          roleRef: {
            kind: ClusterRole.kind,
            apiGroup: 'rbac.authorization.k8s.io',
            name: 'admin',
          },
          subjects: [
            new Subject({
              kind: 'User',
              name: 'admin@example.com',
            }),
          ],
        }),
        Namespace.named('backend-prod'),
        new RoleBinding({
          metadata: {
            name: 'admin',
            namespace: 'backend-prod',
          },
          roleRef: {
            kind: ClusterRole.kind,
            apiGroup: 'rbac.authorization.k8s.io',
            name: 'admin',
          },
          subjects: [
            new Subject({
              kind: 'User',
              name: 'admin@example.com',
            }),
          ],
        }),
      ]),
    ),
  );

  it(
    'expands a Team with both user and group roles',
    RUNNER.run(
      new Configs([
        team('backend', {
          role: 'admin',
          groups: ['sre@example.com'],
          users: ['admin@example.com'],
        }),
      ]),
      new Configs([
        team('backend', {
          role: 'admin',
          groups: ['sre@example.com'],
          users: ['admin@example.com'],
        }),
        Namespace.named('backend-dev'),
        new RoleBinding({
          metadata: {
            name: 'admin',
            namespace: 'backend-dev',
          },
          roleRef: {
            kind: ClusterRole.kind,
            apiGroup: 'rbac.authorization.k8s.io',
            name: 'admin',
          },
          subjects: [
            new Subject({
              kind: 'User',
              name: 'admin@example.com',
            }),
            new Subject({
              kind: 'Group',
              name: 'sre@example.com',
            }),
          ],
        }),
        Namespace.named('backend-prod'),
        new RoleBinding({
          metadata: {
            name: 'admin',
            namespace: 'backend-prod',
          },
          roleRef: {
            kind: ClusterRole.kind,
            apiGroup: 'rbac.authorization.k8s.io',
            name: 'admin',
          },
          subjects: [
            new Subject({
              kind: 'User',
              name: 'admin@example.com',
            }),
            new Subject({
              kind: 'Group',
              name: 'sre@example.com',
            }),
          ],
        }),
      ]),
    ),
  );
});

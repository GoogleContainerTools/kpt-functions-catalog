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

import { Configs, Runner } from 'kpt-functions';
import { isTeam, Team } from './gen/dev.cft.anthos.v1alpha1';
import { Namespace } from './gen/io.k8s.api.core.v1';
import { RoleBinding, Subject } from './gen/io.k8s.api.rbac.v1';

const ENVIRONMENTS = ['dev', 'prod'];

// Generates native K8S resources from Anthos Toolkit Team custom resources.
export function hydrateAnthosTeam(configs: Configs) {
  configs.get(isTeam).forEach((team) => {
    const name = team.metadata.name;

    ENVIRONMENTS.forEach((suffix) => {
      const ns = `${name}-${suffix}`;
      configs.insert(Namespace.named(ns));
      configs.insert(...expandTeam(team, ns));
    });
  });
}

function roleSubjects(item: Team.Spec.Item): Subject[] {
  const userSubjects: Subject[] = (item.users || []).map(
    (user) =>
      new Subject({
        kind: 'User',
        name: user,
      }),
  );
  const groupSubjects: Subject[] = (item.groups || []).map(
    (group) =>
      new Subject({
        kind: 'Group',
        name: group,
      }),
  );
  return userSubjects.concat(groupSubjects);
}

function expandTeam(team: Team, namespace: string): RoleBinding[] {
  return (team.spec.roles || []).map((item) => {
    return new RoleBinding({
      metadata: {
        name: item.role,
        namespace,
      },
      subjects: roleSubjects(item),
      roleRef: {
        kind: 'ClusterRole',
        name: item.role,
        apiGroup: 'rbac.authorization.k8s.io',
      },
    });
  });
}

export const RUNNER = Runner.newFunc(hydrateAnthosTeam);

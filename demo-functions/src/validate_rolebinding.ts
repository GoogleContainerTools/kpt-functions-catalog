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

import { Configs, KubernetesObject, newManifestError, Param, Runner } from 'kpt-functions';
import { isRoleBinding } from './gen/io.k8s.api.rbac.v1';

export const SUBJECT_NAME = new Param('subject_name', {
  required: true,
  help: 'Banned RoleBinding subjects.name',
});

// Validates there is no RBAC RoleBinding with the given subject name.
export function disallowRoleBindingSubject(configs: Configs) {
  const subjectName = configs.getParam(SUBJECT_NAME)!;

  const rbs: KubernetesObject[] = configs
    .get(isRoleBinding)
    .filter((rb) => rb && rb.subjects && rb.subjects.find((s) => s.name === subjectName));

  if (rbs.length) {
    return newManifestError('Found RoleBindings with banned subjects', ...rbs);
  }
  return;
}

export const RUNNER = Runner.newFunc(disallowRoleBindingSubject, SUBJECT_NAME);

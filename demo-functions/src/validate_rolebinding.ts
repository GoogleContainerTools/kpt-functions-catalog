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

import { KubernetesObject, newManifestError, KptFunc } from 'kpt-functions';
import { isRoleBinding } from './gen/io.k8s.api.rbac.v1';

export const SUBJECT_NAME = 'subject_name';

export const disallowRoleBindingSubject: KptFunc = (configs) => {
  const subjectName = configs.getFunctionConfigValueOrThrow(SUBJECT_NAME);

  const rbs: KubernetesObject[] = configs
    .get(isRoleBinding)
    .filter((rb) => rb && rb.subjects && rb.subjects.find((s) => s.name === subjectName));

  if (rbs.length) {
    return newManifestError('Found RoleBindings with banned subjects', ...rbs);
  }
  return;
};

disallowRoleBindingSubject.usage = `
Disallows RBAC RoleBinding objects with the given subject name.

Configured using a ConfigMap with the following key:

${SUBJECT_NAME}: RoleBinding subjects.name to disallow.

Example:

apiVersion: v1
kind: ConfigMap
data:
  ${SUBJECT_NAME}: alice
metadata:
  name: my-config
`;

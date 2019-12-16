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
import { isPodSecurityPolicy } from './gen/io.k8s.api.policy.v1beta1';

// Ensures allowPrivilegeEscalation is always set to false for PodSecurityPolicies.
export function recommendPsp(configs: Configs) {
  configs
    .get(isPodSecurityPolicy)
    .filter((psp) => psp.spec && psp.spec.allowPrivilegeEscalation !== false)
    .forEach((psp) => (psp!.spec!.allowPrivilegeEscalation = false));
}

export const RUNNER = Runner.newFunc(recommendPsp);

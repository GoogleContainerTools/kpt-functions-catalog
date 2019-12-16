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

import * as kpt from 'kpt-functions';
import { isNamespace } from './gen/io.k8s.api.core.v1';

/**
 * Forbids Configs from declaring metadata.namespace or a Namespace's metadata.name that will be
 * rejected by the APIServer.
 *
 * Verifying the existence of metadata.namespace for Namespaced types, or verifying the
 * non-existence for non-Namespaced types is out of the scope of this validator.
 */
export function validateNamespaceName(configs: kpt.Configs) {
  // TODO(b/143533250): Allow ignoring types/specific instances.
  const errs: kpt.ConfigError[] = configs
    .getAll()
    .map(validateNamespace)
    .filter(kpt.isConfigError);

  if (errs.length) {
    // TODO(b/143538989): Allow returning multiple errors.
    return errs[0];
  }
  return;
}

// The regex Kubernetes uses to validate Namespace names.
const NAMESPACE_REGEX = new RegExp('^[a-z0-9]([-a-z0-9]{0,61}[a-z0-9])?$');

function validateNamespace(o: kpt.KubernetesObject): undefined | kpt.ConfigError {
  let namespace: string;
  if (isNamespace(o)) {
    // This is a Namespace, so test its metadata.name.
    namespace = o.metadata.name;
  } else if (o.metadata.namespace) {
    // This object declares metadata.namespace, so validate it.
    namespace = o.metadata.namespace;
  } else {
    // This object does not declare metadata.namespace and it is not a Namespace, so ignore it.
    return;
  }

  if (!NAMESPACE_REGEX.test(namespace) || (isNamespace(o) && o.metadata.name === '')) {
    // A Namespace may not declare metadata.name as empty string, but if an object declares
    // metadata.namespace as empty string it automatically resolves to default.
    if (isNamespace(o)) {
      kpt.newManifestError(
        `A Namespace\'s metadata.name MUST be:
1) nonempty;
2) 63 characters or fewer;
3) consist of lowercase letters (a-z), digits (0-9), and hyphen \`-\`; and
4) begin and end with a lowercase letter or digit`,
        o,
      );
    }
    return kpt.newManifestError(
      `If declared, a Config\'s metadata.namespace MUST be:
1) 63 characters or fewer;
2) consist of lowercase letters (a-z), digits (0-9), and hyphen \`-\`; and
3) begin and end with a lowercase letter or digit`,
      o,
    );
  }

  return;
}

export const RUNNER = kpt.Runner.newFunc(validateNamespaceName);

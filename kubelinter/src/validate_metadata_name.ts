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

/**
 * Forbids KubernetesObjects from declaring metadata.name values that will be rejected by the
 * APIServer.
 */
export function validateMetadataName(configs: kpt.Configs) {
  // TODO(b/143533250): Allow ignoring types/specific instances.
  const errs: kpt.ConfigError[] = configs
    .getAll()
    .map(validateName)
    .filter(kpt.isConfigError);

  if (errs.length) {
    // TODO(b/143538989): Allow returning multiple errors.
    return errs[0];
  }
  return;
}

const NAME_REGEX = new RegExp('^[a-z0-9-.]{1,253}$');

function validateName(o: kpt.KubernetesObject): undefined | kpt.ConfigError {
  const name = o.metadata.name;
  if (!NAME_REGEX.test(name) || name === '.' || name === '..') {
    return kpt.newManifestError(
      'metadata.name MUST be:\n' +
        '1) nonempty;\n' +
        '2) 253 characters or fewer;\n' +
        '3) consist of lowercase letters (a-z), digits (0-9), hyphen `-`, and period `.`; and\n' +
        '4) not be exactly `.` or `..`',
    );
  }
  return;
}

export const RUNNER = kpt.Runner.newFunc(validateMetadataName);

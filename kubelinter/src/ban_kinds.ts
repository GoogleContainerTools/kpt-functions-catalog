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
import * as apps_v1 from './gen/io.k8s.api.apps.v1';
import * as apps_v1beta1 from './gen/io.k8s.api.apps.v1beta1';
import * as authentication_v1 from './gen/io.k8s.api.authentication.v1';
import * as authorization_v1 from './gen/io.k8s.api.authorization.v1';
import * as scaling_v1 from './gen/io.k8s.api.autoscaling.v1';
import * as batch_v1 from './gen/io.k8s.api.batch.v1';
import * as cert_v1beta1 from './gen/io.k8s.api.certificates.v1beta1';
import * as core_v1 from './gen/io.k8s.api.core.v1';
import * as events_v1beta1 from './gen/io.k8s.api.events.v1beta1';
import * as extensions_v1beta1 from './gen/io.k8s.api.extensions.v1beta1';

/**
 * Bans specified Kubernetes resource types.
 *
 * Currently bans obviously bad types to have in version control, such as Secrets and
 * TokenReviews.
 *
 * In the future will allow user-specified bans of additional types, as well as ignoring defaults.
 *
 * @returns a ConfigError if any banned Kinds are present in Configs.
 */
export function banKinds(configs: kpt.Configs) {
  const banDetector = new BanDetector(DEFAULT_BANNED);

  // Detection is amortized O(# objects) + O(# unique types) time complexity.
  const illegalGVKs: kpt.KubernetesObject[][] = configs
    .groupBy((o: kpt.KubernetesObject): string => {
      // Group by GVK since whether an object is banned is only dependent on its GVK.
      return `${o.apiVersion}/${o.kind}`;
    })
    .map((kv: [string, kpt.KubernetesObject[]]): kpt.KubernetesObject[] => {
      // Drop the key since we no longer need it.
      return kv[1];
    })
    .filter((kos: kpt.KubernetesObject[]): boolean => {
      // We only need to check one in each group to know if the entire group is banned.
      return banDetector.isBanned(kos[0]);
    });

  if (illegalGVKs.length > 0) {
    // Writing the error message is O(# banned objects)
    const illegal: kpt.KubernetesObject[] = ([] as kpt.KubernetesObject[]).concat(...illegalGVKs);
    return kpt.newManifestError('Objects with banned kinds', ...illegal);
  }
  return;
}

export const RUNNER = kpt.Runner.newFunc(banKinds);

/**
 * defaultBanned is the set of GVKs which Kubelinter bans of the types which appear in a default
 * Kubernetes 1.14 installation from being declared in repositories.
 *
 * To be safe we ban all Versions for a Group/Kind since they generally all versions share the same
 * problems with being declared.
 *
 * Some Kinds appear in multiple API Groups but are basically the same object because they've moved
 * around during Kubernetes's history. For example, `Scale` has lived in at least three different
 * APIGroups but someone may define a `Scale` CRD that *is* appropriate to declare in a repository.
 * Which APIGroups have "good" and "bad" `Scale` definitions to define in repositories cannot be
 * automatically determined, so an entry from each Group must be manually specified below.
 */
// TODO(b/143533434): Error messages should indicate why the type is banned.
const DEFAULT_BANNED: Set<string> = new Set([
  // Immutable data snapshots
  unversioned(apps_v1.ControllerRevision),

  // Single use requests for permissions
  unversioned(authentication_v1.TokenReview),

  unversioned(authorization_v1.LocalSubjectAccessReview),
  unversioned(authorization_v1.SelfSubjectAccessReview),
  unversioned(authorization_v1.SubjectAccessReview),

  unversioned(cert_v1beta1.CertificateSigningRequest),

  // Single-use imperative objects
  unversioned(batch_v1.Job),
  unversioned(core_v1.Event),
  unversioned(events_v1beta1.Event),

  unversioned(apps_v1beta1.Scale),
  unversioned(scaling_v1.Scale),
  unversioned(extensions_v1beta1.Scale),

  // Don't declare credentials
  unversioned(core_v1.Secret),

  // Deprecated
  unversioned(core_v1.Binding),

  // Status objects
  unversioned(core_v1.ComponentStatus),
  unversioned(core_v1.Endpoints),
]);

interface GroupKind {
  group: string;
  kind: string;
}

// Bans all versions of a Group/Kind combination.
function unversioned(gk: GroupKind): string {
  if (gk.group === '') {
    return `*/${gk.kind}`;
  }
  return `${gk.group}/*/${gk.kind}`;
}

class BanDetector {
  isBanned(o: kpt.KubernetesObject): boolean {
    // TODO(b/143533250): Add directive(s) for ignoring for specific objects/types.
    // TODO(b/143533250): Allow adding additional types to ban.
    // TODO(b/143533250): Allow banning specific versions.
    // TODO(b/143533250): Allow banning all Kinds in a Group/Version or Group.
    // TODO(b/143533250): Allow banning a Kind name, regardless of Group/Version.
    // Check for unversioned bans.
    const gvk = groupVersionKind(o);
    return this.banned.has(unversioned(gvk));
  }

  constructor(private banned: Set<string>) {}
}

// TODO(b/143533151): Move everything below to kpt-functions, comment, and test.
//  Quality/comments/testing of this code is out of scope of this CL.
export interface APIVersionKind {
  apiVersion: string;
  kind: string;
}

interface GroupVersionKind {
  group: string;
  version: string;
  kind: string;
}

function groupVersion(apiVersion: string): [string, string] {
  const gv = apiVersion.split('/');
  if (gv.length === 1) {
    return ['', gv[0]];
  }
  return [gv[0], gv[1]];
}

function groupVersionKind(o: APIVersionKind): GroupVersionKind {
  const [group, version]: [string, string] = groupVersion(o.apiVersion);
  return {
    group: group,
    version: version,
    kind: o.kind,
  };
}

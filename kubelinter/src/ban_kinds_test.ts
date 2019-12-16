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

import { ConfigError, Configs, KubernetesObject, TestRunner } from 'kpt-functions';
import { APIVersionKind, banKinds } from './ban_kinds';
import * as apps_v1 from './gen/io.k8s.api.apps.v1';
import * as apps_v1beta1 from './gen/io.k8s.api.apps.v1beta1';
import * as apps_v1beta2 from './gen/io.k8s.api.apps.v1beta2';
import * as authentication_v1 from './gen/io.k8s.api.authentication.v1';
import * as authentication_v1beta1 from './gen/io.k8s.api.authentication.v1beta1';
import * as authorization_v1 from './gen/io.k8s.api.authorization.v1';
import * as authorization_v1beta1 from './gen/io.k8s.api.authorization.v1beta1';
import * as scaling_v1 from './gen/io.k8s.api.autoscaling.v1';
import * as batch_v1 from './gen/io.k8s.api.batch.v1';
import * as cert_v1beta1 from './gen/io.k8s.api.certificates.v1beta1';
import * as core_v1 from './gen/io.k8s.api.core.v1';
import * as events_v1beta1 from './gen/io.k8s.api.events.v1beta1';
import * as extensions_v1beta1 from './gen/io.k8s.api.extensions.v1beta1';

const TEST_RUNNER = new TestRunner(banKinds);

function fake(...avks: APIVersionKind[]): Configs {
  return new Configs(
    avks.map(
      (avk): KubernetesObject => {
        return {
          apiVersion: avk.apiVersion,
          kind: avk.kind,
          metadata: {
            name: 'foo',
          },
        };
      },
    ),
  );
}

describe(banKinds.name, () => {
  it('passes empty configs', TEST_RUNNER.run());

  it('does not ban Namespaces', TEST_RUNNER.run(fake(core_v1.Namespace)));

  // Exhaustively test all versions of banned types.
  // Immutable data snapshots
  ensureBanned(apps_v1.ControllerRevision);
  ensureBanned(apps_v1beta1.ControllerRevision);
  ensureBanned(apps_v1beta2.ControllerRevision);

  // Single use requests for permissions
  ensureBanned(authentication_v1.TokenReview);
  ensureBanned(authentication_v1beta1.TokenReview);

  ensureBanned(authorization_v1.LocalSubjectAccessReview);
  ensureBanned(authorization_v1beta1.LocalSubjectAccessReview);
  ensureBanned(authorization_v1.SelfSubjectAccessReview);
  ensureBanned(authorization_v1beta1.SelfSubjectAccessReview);
  ensureBanned(authorization_v1.SubjectAccessReview);
  ensureBanned(authorization_v1beta1.SubjectAccessReview);

  ensureBanned(cert_v1beta1.CertificateSigningRequest);

  // Single-use imperative objects
  ensureBanned(batch_v1.Job);

  ensureBanned(core_v1.Event);
  ensureBanned(events_v1beta1.Event);

  ensureBanned(apps_v1beta1.Scale);
  ensureBanned(apps_v1beta2.Scale);
  ensureBanned(extensions_v1beta1.Scale);
  ensureBanned(scaling_v1.Scale);

  // Don't declare credentials
  ensureBanned(core_v1.Secret);

  // Deprecated types
  ensureBanned(core_v1.Binding);

  // Status objects
  ensureBanned(core_v1.ComponentStatus);
  ensureBanned(core_v1.Endpoints);

  it(
    'rejects if good + banned',
    TEST_RUNNER.run(fake(core_v1.Namespace, core_v1.Secret), new ConfigError('')),
  );

  it('passes if good + good', TEST_RUNNER.run(fake(core_v1.Namespace, core_v1.Pod)));
});

// TODO: Test that the banned object is referenced in the error.
function ensureBanned(avk: APIVersionKind): void {
  // TODO: Test that error actually references object.
  it(`bans ${avk.apiVersion}.${avk.kind}`, TEST_RUNNER.run(fake(avk), new ConfigError('')));
}

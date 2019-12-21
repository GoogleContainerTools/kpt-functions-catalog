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

import { Configs } from '@googlecontainertools/kpt-functions';
import { noOp } from './no_op';
import { Role } from './gen/io.k8s.api.rbac.v1';

describe('noOp', () => {
  it('empty', () => {
    const configs = new Configs();
    noOp(configs);

    expect(configs).toEqual(new Configs());
  });

  it('pass through', () => {
    const role = Role.named('alice');
    const configs = new Configs([role]);

    noOp(configs);

    expect(configs.getAll()).toEqual([role]);
  });
});

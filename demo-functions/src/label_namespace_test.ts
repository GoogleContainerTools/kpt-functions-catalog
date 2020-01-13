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
import { labelNamespace, LABEL_NAME, LABEL_VALUE } from './label_namespace';
import { Namespace, ConfigMap } from './gen/io.k8s.api.core.v1';

const TEST_NAMESPACE = 'testNamespace';
const TEST_LABEL_NAME = 'costCenter';
const TEST_LABEL_VALUE = 'xyz';

describe('labelNamespace', () => {
  let functionConfig = ConfigMap.named('foo');
  functionConfig.data = {};
  functionConfig.data[LABEL_NAME] = TEST_LABEL_NAME;
  functionConfig.data[LABEL_VALUE] = TEST_LABEL_VALUE;

  it('empty input ok', () => {
    expect(labelNamespace(new Configs(undefined, functionConfig))).toBeUndefined();
  });

  it('adds label namespace when metadata.labels is undefined', () => {
    const actual = new Configs(undefined, functionConfig);
    actual.insert(Namespace.named(TEST_NAMESPACE));

    labelNamespace(actual);

    const expected = new Configs();
    expected.insert(
      new Namespace({
        metadata: {
          name: TEST_NAMESPACE,
          labels: { [TEST_LABEL_NAME]: TEST_LABEL_VALUE },
        },
      }),
    );

    expect(actual.getAll()).toEqual(expected.getAll());
  });

  it('adds label to namespace when metadata.labels is defined', () => {
    const actual = new Configs(undefined, functionConfig);
    actual.insert(
      new Namespace({
        metadata: {
          name: TEST_NAMESPACE,
          labels: { a: 'b' },
        },
      }),
    );

    const expected = new Configs();
    expected.insert(
      new Namespace({
        metadata: {
          name: TEST_NAMESPACE,
          labels: {
            a: 'b',
            [TEST_LABEL_NAME]: TEST_LABEL_VALUE,
          },
        },
      }),
    );

    labelNamespace(actual);

    expect(actual.getAll()).toEqual(expected.getAll());
  });
});

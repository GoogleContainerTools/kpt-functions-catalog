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

import { Configs } from 'kpt-functions';
import { addLabelToAllNamespaces, LABEL_NAME, LABEL_VALUE } from './add_label_to_all_namespaces';
import { Namespace } from './gen/io.k8s.api.core.v1';

const TEST_NAMESPACE = 'testNamespace';
const TEST_LABEL_NAME = 'costCenter';
const TEST_LABEL_VALUE = 'xyz';

describe('addLabelToAllNamespaces', () => {
  const params = new Map([
    [LABEL_NAME.name, TEST_LABEL_NAME],
    [LABEL_VALUE.name, TEST_LABEL_VALUE],
  ]);

  it('empty input', () => {
    expect(addLabelToAllNamespaces(new Configs(undefined, params))).toBeUndefined();
  });

  it('adds label namespace is metadata.labels is undefined', () => {
    const actual = new Configs(undefined, params);
    actual.insert(Namespace.named(TEST_NAMESPACE));

    addLabelToAllNamespaces(actual);

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

  it('adds label to namespace if metadata.labels is defined', () => {
    const actual = new Configs(undefined, params);
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

    addLabelToAllNamespaces(actual);

    expect(actual.getAll()).toEqual(expected.getAll());
  });
});

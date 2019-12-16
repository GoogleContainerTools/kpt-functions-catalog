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
import { Namespace } from './gen/io.k8s.api.core.v1';
import { validateNamespaceName } from './validate_namespace_name';

const RUNNER = new TestRunner(validateNamespaceName);

function theNamespace(namespace: string | undefined): TestCase {
  return new TestCase(namespace);
}

class TestCase {
  constructor(private namespace: string | undefined) {}

  isValid() {
    this.isValidMetadataNamespace();
    this.isValidNamespace();
  }

  isInvalid() {
    this.isInvalidMetadataNamespace();
    this.isInvalidNamespace();
  }

  isValidMetadataNamespace() {
    const obj = objectWithNamespace(this.namespace);
    it(`valid metadata.namespace: ${this.namespace}`, RUNNER.run(new Configs([obj])));
  }

  isInvalidMetadataNamespace() {
    const obj = objectWithNamespace(this.namespace);
    it(
      `valid metadata.namespace: ${this.namespace}`,
      RUNNER.run(new Configs([obj]), new ConfigError('')),
    );
  }

  isValidNamespace() {
    // Impossible for metadata.name to be undefined.
    const ns = Namespace.named(this.namespace!);
    it(`valid Namespace metadata.name: ${this.namespace}`, RUNNER.run(new Configs([ns])));
  }

  isInvalidNamespace() {
    // Impossible for metadata.name to be undefined.
    const ns = Namespace.named(this.namespace!);
    it(
      `valid Namespace metadata.name: ${this.namespace}`,
      RUNNER.run(new Configs([ns]), new ConfigError('')),
    );
  }
}

function objectWithNamespace(namespace: string | undefined): KubernetesObject {
  return {
    apiVersion: 'v1',
    kind: 'Object',
    metadata: {
      name: 'foo',
      namespace: namespace,
    },
  };
}

describe('validateNamespaceName', () => {
  it('passes empty configs', RUNNER.run());

  // Verify character types
  theNamespace('aaa').isValid();
  theNamespace('a0a').isValid();
  theNamespace('a-a').isValid();
  theNamespace('a.a').isInvalid();
  theNamespace('aAa').isInvalid();

  // Verify special cases
  theNamespace('').isValidMetadataNamespace();
  theNamespace('').isInvalidNamespace();
  theNamespace(undefined).isValidMetadataNamespace();
  theNamespace('a').isValid();
  theNamespace('a-').isInvalid();
  theNamespace('-a').isInvalid();

  // Verify longer names
  theNamespace('a-b-c').isValid();
  theNamespace('abcd.efg').isInvalid();

  // Verify length limit
  theNamespace('a'.repeat(63)).isValid();
  theNamespace('a'.repeat(64)).isInvalid();

  it(
    'passes two good configs',
    RUNNER.run(new Configs([objectWithNamespace('aaa'), objectWithNamespace('bbb')])),
  );
  it(
    'fails one good one bad',
    RUNNER.run(
      new Configs([objectWithNamespace('aaa'), objectWithNamespace('BBB')]),
      new ConfigError(''),
    ),
  );
});

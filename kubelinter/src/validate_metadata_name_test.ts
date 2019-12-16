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
import { validateMetadataName } from './validate_metadata_name';

const RUNNER = new TestRunner(validateMetadataName);

describe('validateMetadataName', () => {
  it('passes empty configs', RUNNER.run());

  // Verify character types
  accepts('aaa');
  accepts('0aa');
  accepts('-aa');
  rejects('Aaa');
  rejects('/aa');

  // Verify special cases
  rejects('');
  rejects('.');
  rejects('..');
  accepts('...');

  // Verify longer names
  accepts('a-b.c');
  rejects('abcd%efg');

  // Verify length limit
  accepts('a'.repeat(253));
  rejects('a'.repeat(254));

  it('passes two good configs', RUNNER.run(new Configs([objectNamed('aaa'), objectNamed('bbb')])));
  it(
    'fails one good one bad',
    RUNNER.run(new Configs([objectNamed('aaa'), objectNamed('BBB')]), new ConfigError('')),
  );
});

function objectNamed(name: string): KubernetesObject {
  return {
    apiVersion: 'v1',
    kind: 'Object',
    metadata: {
      name: name,
    },
  };
}

function rejects(name: string) {
  const obj = objectNamed(name);
  it(`rejects '${name}'`, RUNNER.run(new Configs([obj]), new ConfigError('')));
}

function accepts(name: string) {
  const obj = objectNamed(name);
  it(`accepts '${name}'`, RUNNER.run(new Configs([obj])));
}

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

import { compareSync } from 'dir-compare';
import * as fs from 'fs';
import { Configs, FileFormat, isConfigError, readConfigs, writeConfigs } from 'kpt-functions';
import * as os from 'os';
import * as path from 'path';
import { readYAMLDir, SOURCE_DIR } from './source_yaml_dir';

describe('readYAMLDir', () => {
  it('works on empty dir', () => {
    const sourceDir = path.resolve(__dirname, '../test-data/source/empty');

    const configs = new Configs(undefined, new Map([[SOURCE_DIR.name, sourceDir]]));
    readYAMLDir(configs);
    if (isConfigError(configs)) {
      fail('Unexpected error: ' + configs);
    } else {
      expect(configs.getAll().length).toBe(0);
    }
  });

  it('replicates test dir', () => {
    const sourceDir = path.resolve(__dirname, '../test-data/source/foo-yaml');
    const expectedIntermediateFile = path.resolve(__dirname, '../test-data/intermediate/foo.yaml');
    const expectedConfigs = readConfigs(expectedIntermediateFile, FileFormat.YAML);

    const actualConfigs = new Configs(undefined, new Map([[SOURCE_DIR.name, sourceDir]]));
    readYAMLDir(actualConfigs);

    expect(actualConfigs.getAll()).toEqual(expectedConfigs.getAll());

    const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'yaml-dir-source-test'));
    const intermediateFile = path.join(tmpDir, 'foo.yaml');
    writeConfigs(intermediateFile, actualConfigs, FileFormat.YAML);
    const res = compareSync(expectedIntermediateFile, intermediateFile, {
      compareContent: true,
    });
    if (res.differences) {
      console.log(res.diffSet);
      fail('Found differences between actual and generated intermediate files.');
    }
  });

  it('fails for invalid KubernetesObjects', () => {
    const sourceDir = path.resolve(__dirname, '../test-data/source/invalid');

    const actualConfigs = new Configs(undefined, new Map([[SOURCE_DIR.name, sourceDir]]));
    const err = readYAMLDir(actualConfigs);
    if (!isConfigError(err)) {
      fail('Expected error, but got configs: ' + JSON.stringify(actualConfigs, undefined, 2));
    }
  });
});

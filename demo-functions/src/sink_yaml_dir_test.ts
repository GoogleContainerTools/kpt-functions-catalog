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
import * as fs from 'fs-extra';
import * as kpt from '@googlecontainertools/kpt-functions';
import * as os from 'os';
import * as path from 'path';
import { Namespace, Pod, ConfigMap } from './gen/io.k8s.api.core.v1';
import { Role, RoleBinding } from './gen/io.k8s.api.rbac.v1';
import { buildSourcePath, OVERWRITE, SINK_DIR, writeYAMLDir } from './sink_yaml_dir';

const INTERMEDIATE_FILE = path.resolve(__dirname, '..', 'test-data', 'intermediate', 'foo.yaml');
const SINK_DIR_EXPECTED = path.resolve(__dirname, '..', 'test-data', 'sink', 'foo-yaml');

function readIntermediate(): kpt.Configs {
  return kpt.readConfigs(INTERMEDIATE_FILE, kpt.FileFormat.YAML);
}

function matchesExpected(dir: string) {
  const res = compareSync(dir, SINK_DIR_EXPECTED, {
    compareContent: true,
  });
  if (res.differences) {
    console.log(res.diffSet);
    fail('Found differences between actual and generated directories');
  }
}

describe('writeYAMLDir', () => {
  let tmpDir: string = '';
  let functionConfig = ConfigMap.named('config');

  beforeEach(() => {
    // Ensures tmpDir is unset before testing. Detects incorrectly running tests in parallel, or
    // tests not cleaning up properly.
    expect(tmpDir).toEqual('');
    tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'yaml-dir-sink-test'));
    functionConfig.data = {};
  });

  afterEach(() => {
    // Remove tmpDir so no other tests can have access to the data.
    fs.removeSync(tmpDir);
    // Reset tmpDir value to confirm test finished normally.
    tmpDir = '';
  });

  it('test dir', () => {
    const input = readIntermediate();
    functionConfig.data![SINK_DIR] = tmpDir;
    const configs = new kpt.Configs(input.getAll(), functionConfig);

    writeYAMLDir(configs);

    matchesExpected(tmpDir);
  });

  it("throws if --overwrite isn't passed for non-empty directory", () => {
    fs.copySync(SINK_DIR_EXPECTED, tmpDir);
    const input = readIntermediate();
    functionConfig.data![SINK_DIR] = tmpDir;
    const configs = new kpt.Configs(input.getAll(), functionConfig);

    expect(() => writeYAMLDir(configs)).toThrow();
  });

  it("silently makes output directory if it doesn't exist", () => {
    const sinkDir = path.resolve(tmpDir, 'foo');
    const input = readIntermediate();
    functionConfig.data![SINK_DIR] = sinkDir;
    const configs = new kpt.Configs(input.getAll(), functionConfig);

    writeYAMLDir(configs);

    matchesExpected(sinkDir);
  });

  it('overwrites if --overwrite is passed for non-empty directory', () => {
    fs.copySync(SINK_DIR_EXPECTED, tmpDir);
    // Modify contents of existing file.
    fs.copySync(
      path.resolve(tmpDir, 'foo-corp-1.0.0', 'podsecuritypolicy_psp.yaml'),
      path.resolve(tmpDir, 'foo-corp-1.0.0', 'clusterrole_pod-creator.yaml'),
      { overwrite: true },
    );
    // Move file to a different location.
    fs.moveSync(
      path.resolve(tmpDir, 'foo-corp-1.0.0', 'podsecuritypolicy_psp.yaml'),
      path.resolve(tmpDir, 'foo-corp-1.0.0', 'other.yaml'),
    );
    const input = readIntermediate();
    functionConfig.data![SINK_DIR] = tmpDir;
    functionConfig.data![OVERWRITE] = 'true';
    const configs = new kpt.Configs(input.getAll(), functionConfig);

    writeYAMLDir(configs);

    // Ensure the resulting directory is actually overwritten.
    matchesExpected(tmpDir);
  });
});

describe('buildSourcePath', () => {
  it('reads the source path annotation if set', () => {
    const o = Namespace.named('qux');
    kpt.addAnnotation(o, kpt.SOURCE_PATH_ANNOTATION, 'file.yaml');

    const result = buildSourcePath(o);
    expect(result).toEqual('file.yaml');
  });

  it('puts non-namespaced objects in the top-level directory', () => {
    // This test is intentionally using RoleBinding, a Namespaced resource, to highlight that we
    // don't yet know the distinction between Namespaced and non-Namespaced resources except by
    // checking metadata.namespace manually.
    const o = new RoleBinding({
      metadata: {
        name: 'dev',
      },
      roleRef: {
        kind: Role.kind,
        apiGroup: Role.apiVersion.split('/')[0],
        name: 'dev',
      },
    });

    const result = buildSourcePath(o);
    expect(result).toEqual('rolebinding_dev.yaml');
  });

  it("puts namespaced objects in the namespace's directory", () => {
    const o = new Pod({
      metadata: {
        name: 'syncer',
        namespace: 'frontend',
      },
    });

    const result = buildSourcePath(o);
    expect(result).toEqual('frontend/pod_syncer.yaml');
  });

  it('translates empty string to the default Namespace', () => {
    const o = new Role({
      metadata: {
        name: 'pod-editor',
        namespace: '',
      },
    });

    const result = buildSourcePath(o);
    expect(result).toEqual('default/role_pod-editor.yaml');
  });

  it('puts Namespaces in their directory', () => {
    const o = Namespace.named('backend');

    const result = buildSourcePath(o);
    expect(result).toEqual('backend/namespace.yaml');
  });
});

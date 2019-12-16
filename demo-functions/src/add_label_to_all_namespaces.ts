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

import { Configs, Param, Runner } from 'kpt-functions';
import { isNamespace, Namespace } from './gen/io.k8s.api.core.v1';

export const LABEL_NAME = new Param('lable_name', {
  required: true,
  help: 'Label name to annotate namespaces with',
});

export const LABEL_VALUE = new Param('label_value', {
  required: true,
  help: 'Label value to annotate namespaces with',
});

/**
 * Add a label to all namespaces found in configs.
 *
 * @param configs The configs to validate/mutate.
 */
export function addLabelToAllNamespaces(configs: Configs) {
  const labelName = configs.getParam(LABEL_NAME)!;
  const labelValue = configs.getParam(LABEL_VALUE)!;

  configs.get(isNamespace).forEach((n) => addLabelToNamespace(n, labelName, labelValue));
}

function addLabelToNamespace(namespace: Namespace, labelName: string, labelValue: string) {
  if (namespace.metadata.labels === undefined) {
    namespace.metadata.labels = {};
  }

  namespace.metadata.labels[labelName] = labelValue;
}

export const RUNNER = Runner.newFunc(addLabelToAllNamespaces, LABEL_NAME, LABEL_VALUE);

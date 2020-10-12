#!/usr/bin/env bash
#
# Copyright 2020 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# E2E tests for annotate_config.

set -eo pipefail
DIR="$(dirname "$0")"
# shellcheck source=tests/common.sh
source "$DIR"/common.sh

testcase "kpt_annotate_config_imperative"
kpt pkg get "$SDK_REPO"/example-configs example-configs
kpt fn source example-configs |
  kpt fn run --image gcr.io/kpt-functions/annotate-config:"${TAG}" -- annotation_name=configmanagement.gke.io/namespace-selector annotation_value=sre-supported |
  kpt fn sink example-configs
assert_contains_string example-configs/gatekeeper.yaml "configmanagement.gke.io/namespace-selector: sre-supported"

testcase "kpt_annotate_config_declarative"
kpt pkg get "$SDK_REPO"/example-configs example-configs
cat >fc.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  annotations:
    config.k8s.io/function: |
      container:
        image: gcr.io/kpt-functions/annotate-config:${TAG}
    config.kubernetes.io/local-config: 'true'
data:
  annotation_name: configmanagement.gke.io/namespace-selector
  annotation_value: sre-supported
EOF
kpt fn source example-configs |
  kpt fn run --fn-path fc.yaml |
  kpt fn sink example-configs
assert_contains_string example-configs/gatekeeper.yaml "configmanagement.gke.io/namespace-selector: sre-supported"

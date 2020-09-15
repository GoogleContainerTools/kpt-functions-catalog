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
# E2E tests for suggest_psp.

set -eo pipefail
DIR="$(dirname "$0")"
# shellcheck source=tests/common.sh
source "$DIR"/common.sh

testcase "kpt_suggest_psp_imperative"
kpt pkg get "$SDK_REPO"/example-configs example-configs
kpt fn source example-configs |
  kpt fn run --image gcr.io/kpt-functions/suggest-psp:"${TAG}" 2>results.err |
  kpt fn sink example-configs
assert_contains_string results.err "Suggest explicitly disabling privilege escalation"

testcase "kpt_suggest_psp_declarative"
kpt pkg get "$SDK_REPO"/example-configs example-configs
cat >fc.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  annotations:
    config.k8s.io/function: |
      container:
        image: gcr.io/kpt-functions/suggest-psp:${TAG}
    config.kubernetes.io/local-config: 'true'
EOF
kpt fn source example-configs |
  kpt fn run --fn-path fc.yaml 2>results.err |
  kpt fn sink example-configs
assert_contains_string results.err "Suggest explicitly disabling privilege escalation"

testcase "kpt_suggest_psp_declarative_example"
kpt pkg get "$CATALOG_REPO"/examples/suggest-psp .
kpt fn run suggest-psp --results-dir "$(pwd)" || true
assert_contains_string results-0.yaml "Suggest explicitly disabling privilege escalation"

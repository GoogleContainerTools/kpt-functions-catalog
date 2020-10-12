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
# E2E tests for templater.

set -eo pipefail
DIR="$(dirname "$0")"
# shellcheck source=tests/common.sh
source "$DIR"/common.sh

############################
# kpt fn Tests
############################

testcase "kpt_templater_imperative"
kpt fn run . --env TESTTEMPLATERENV="testval" --image gcr.io/kpt-functions/templater:"${TAG}" -- entrypoint="apiVersion: v1
kind: ConfigMap
metadata:
  name: testcfg
data:
  value: {{env \"TESTTEMPLATERENV\" }}"
assert_contains_string configmap_testcfg.yaml "value: testval"

testcase "kpt_templater_declarative"
cat >fc.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  annotations:
    config.k8s.io/function: |
      container:
        image: gcr.io/kpt-functions/templater:${TAG}
        envs:
        - TESTTEMPLATERENV=testval
    config.kubernetes.io/local-config: 'true'
data:
  entrypoint: |
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: testcfg
    data:
      value: {{env "TESTTEMPLATERENV" }}
EOF
kpt fn run .
assert_contains_string configmap_testcfg.yaml "value: testval"

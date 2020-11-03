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
# E2E tests for env-annotation.

set -eo pipefail
DIR="$(dirname "$0")"
# shellcheck source=tests/common.sh
source "$DIR"/common.sh

############################
# kpt fn Tests
############################

testcase "env_annotation_imperative"
cat >configmap_testcfg.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: testcfg
EOF
kpt fn run . --env TESTENV="testval" --image gcr.io/kpt-functions/env-annotation:"${TAG}" -- TESTENV=""
assert_contains_string configmap_testcfg.yaml "TESTENV: 'testval'"

testcase "env_annotation_declarative"
cat >configmap_testcfg.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: testcfg
EOF
cat >fc.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  annotations:
    config.k8s.io/function: |
      container:
        image: gcr.io/kpt-functions/env-annotation:${TAG}
        envs:
        - TESTENV=testval
    config.kubernetes.io/local-config: 'true'
data:
  TESTENV: ''
EOF
kpt fn run .
assert_contains_string configmap_testcfg.yaml "TESTENV: 'testval"

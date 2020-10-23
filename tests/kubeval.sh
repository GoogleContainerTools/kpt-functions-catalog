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
# E2E tests for kubeval.

set -eo pipefail
DIR="$(dirname "$0")"
# shellcheck source=tests/common.sh
source "$DIR"/common.sh

testcase "kpt_kubeval_imperative_no_errors"
kpt pkg get https://github.com/instrumenta/kubeval.git/fixtures .
kpt fn source fixtures/valid* |
  kpt fn run --network --image gcr.io/kpt-functions/kubeval:"${TAG}" >out.yaml
if grep -q "results" out.yaml; then
  fail "Validation error found using kubeval fixtures valid config: " "$(< out.yaml)"
fi

testcase "kpt_kubeval_declarative_no_errors"
kpt pkg get https://github.com/instrumenta/kubeval.git/fixtures .
cat >fc.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  annotations:
    config.k8s.io/function: |
      container:
        image: gcr.io/kpt-functions/kubeval:${TAG}
        network: true
    config.kubernetes.io/local-config: 'true'
EOF
kpt fn source fixtures/valid* |
  kpt fn run --fn-path fc.yaml --network >out.yaml
if grep -q "results" out.yaml; then
  fail "Validation error found using kubeval fixtures valid config: " "$(< out.yaml)"
fi

testcase "kpt_kubeval_imperative_finds_errors"
kpt pkg get https://github.com/instrumenta/kubeval.git/fixtures .
kpt fn source fixtures/*invalid.yaml |
  kpt fn run --network --image gcr.io/kpt-functions/kubeval:"${TAG}" 2>error.txt || true
assert_contains_string error.txt "Invalid type. Expected: \[integer,null\], given: string"

testcase "kpt_kubeval_declarative_finds_errors"
kpt pkg get https://github.com/instrumenta/kubeval.git/fixtures .
cat >fc.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  annotations:
    config.k8s.io/function: |
      container:
        image: gcr.io/kpt-functions/kubeval:${TAG}
        network: true
    config.kubernetes.io/local-config: 'true'
EOF
kpt fn source fixtures/*invalid.yaml |
  kpt fn run --fn-path fc.yaml --network 2>error.txt || true
assert_contains_string error.txt "Invalid type. Expected: \[integer,null\], given: string"

testcase "kpt_kubeval_declarative_example"
kpt pkg get "$CATALOG_REPO"/examples/kubeval .
kpt fn run kubeval --network 2>err.txt || true
assert_contains_string err.txt "Invalid type. Expected: \[integer,null\], given: string"

# TODO: Remove after fixing https://github.com/GoogleContainerTools/kpt/issues/983
testcase "kpt_kubeval_declarative_results"
kpt pkg get "$CATALOG_REPO"/examples/kubeval .
kpt fn run kubeval --network --results-dir /tmp || true
assert_contains_string /tmp/results-0.yaml "path: spec.replicas"

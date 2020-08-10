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
# E2E tests for istioctl_analyze.

set -eo pipefail
DIR="$(dirname "$0")"
# shellcheck source=tests/common.sh
source "$DIR"/common.sh

############################
# Docker Tests
############################
[[ -z "${NODOCKER}" ]] || {
  echo "Skipping docker tests"
  exit 0
}

# TODO: Convert these cases to kpt after fixing the following
# https://github.com/GoogleContainerTools/kpt/issues/823
# https://github.com/GoogleContainerTools/kpt/issues/824
testcase "docker_istioctl_analyze_imperative_no_errors"
kpt pkg get https://github.com/istio/istio.git/samples/addons .
kpt fn source addons |
  docker run -i -u "$(id -u)" gcr.io/kpt-functions/istioctl-analyze:"${TAG}" -d--use-kube=false 2>error.txt |
  kpt fn sink addons
if [ -s error.txt ]; then
  fail "Validation error found using istio addons sample: " "$(< error.txt)"
fi

testcase "docker_istioctl_analyze_imperative_find_errors"
kpt pkg get https://github.com/istio/istio.git/galley/pkg/config/analysis/analyzers/testdata .
kpt fn source testdata |
  docker run -u "$(id -u)" -i gcr.io/kpt-functions/istioctl-analyze:"${TAG}" -d--use-kube=false >out.yaml || true
assert_contains_string out.yaml "Referenced gateway not found"

############################
# kpt fn Tests
############################

testcase "kpt_istioctl_analyze_declarative_no_errors"
kpt pkg get https://github.com/istio/istio.git/samples/addons .
cat >fc.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  annotations:
    config.k8s.io/function: |
      container:
        image: gcr.io/kpt-functions/istioctl-analyze:${TAG}
    config.kubernetes.io/local-config: 'true'
data:
  "flags": [ "--recursive" ]
  "--use-kube": "false"
EOF
kpt fn source addons |
  kpt fn run --fn-path fc.yaml 2>error.txt |
  kpt fn sink addons
if [ -s error.txt ]; then
  fail "Validation error found using istio addons sample: " "$(< error.txt)"
fi

testcase "kpt_istioctl_analyze_declarative_find_errors"
kpt pkg get https://github.com/istio/istio.git/galley/pkg/config/analysis/analyzers/testdata .
cat >fc.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  annotations:
    config.k8s.io/function: |
      container:
        image: gcr.io/kpt-functions/istioctl-analyze:${TAG}
    config.kubernetes.io/local-config: 'true'
data:
  "flags": [ "--recursive" ]
  "--use-kube": "false"
EOF
kpt fn run testdata --fn-path fc.yaml 2>error.txt || true
assert_contains_string error.txt "Referenced selector not found"

testcase "kpt_istioctl_analyze_declarative_example"
kpt pkg get https://github.com/prachirp/kpt-functions-catalog.git/examples/istioctl-analyze@istioctl-blueprint .
kpt fn run istioctl-analyze/configs --fn-path istioctl-analyze/functions --results-dir istioctl-analyze/results
assert_contains_string istioctl-analyze/results/results-0.yaml "Port name  (port: 5000, targetPort: 0) doesn''t follow the naming convention"

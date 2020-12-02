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
# E2E tests for set_namespace_go.

set -eo pipefail
DIR="$(dirname "$0")"
# shellcheck source=tests/common.sh
source "$DIR"/common.sh

testcase "kpt_set_namespace_go_docker_imperative"
kpt pkg get "$SDK_REPO"/example-configs example-configs
kpt fn run . --image gcr.io/kpt-functions/set-namespace:"${TAG}" -- namespace=example-ns
assert_contains_string example-configs/gatekeeper.yaml "namespace: example-ns"

testcase "kpt_set_namespace_go_docker_declarative"
kpt pkg get "$SDK_REPO"/example-configs example-configs
cat >fc.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  annotations:
    config.k8s.io/function: |
      container:
        image: gcr.io/kpt-functions/set-namespace:${TAG}
    config.kubernetes.io/local-config: 'true'
data:
  "namespace": "example-ns"
EOF
kpt fn source example-configs |
  kpt fn run --fn-path fc.yaml >out.yaml
assert_contains_string out.yaml "namespace: example-ns"

testcase "kpt_set_namespace_go_declarative_example"
kpt pkg get "$CATALOG_REPO"/examples/set-namespace .
kpt fn run set-namespace/configs --fn-path set-namespace/functions
assert_contains_string set-namespace/configs/example-config.yaml "namespace: example-ns"

testcase "kpt_set_namespace_go_docker_imperative_idempotency"
kpt pkg get "$SDK_REPO"/example-configs example-configs
# Use example-configs/shipping-dev because after the first time of applying namespace,
# there will be some resources with duplicate IDs. This case is not allowed.
kpt fn run example-configs/shipping-dev --image gcr.io/kpt-functions/set-namespace:"${TAG}" -- namespace=example-ns
assert_contains_string example-configs/shipping-dev/rolebinding_viewers.yaml "namespace: example-ns"
cp example-configs/shipping-dev/rolebinding_viewers.yaml example-configs/shipping-dev/rolebinding_viewers.yaml.orig
# run function agian with same parameters
kpt fn run example-configs/shipping-dev --image gcr.io/kpt-functions/set-namespace:"${TAG}" -- namespace=example-ns
assert_files_same example-configs/shipping-dev/rolebinding_viewers.yaml.orig example-configs/shipping-dev/rolebinding_viewers.yaml

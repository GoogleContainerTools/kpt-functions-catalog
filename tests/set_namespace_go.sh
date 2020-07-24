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

############################
# kpt fn Tests
############################

# Use the macos executable to test using kpt exec runtime
if [ -n "${NODOCKER}" ]
then
testcase "kpt_set_namespace_go_exec_imperative_success_macos"
kpt pkg get "$SDK_REPO"/example-configs example-configs
kpt pkg get https://github.com/prachirp/kpt-functions-catalog/functions/go@e2e-revamp ./
kpt fn run example-configs --enable-exec --exec-path "$(pwd)"/go/set-namespace/set-namespace-macos -- namespace=example-ns
assert_contains_string example-configs/gatekeeper.yaml "namespace: example-ns"
fi

[[ -z "${NODOCKER}" ]] || {
  echo "Skipping docker tests"
  exit 0
}

testcase "kpt_set_namespace_go_exec_imperative_success_linux"
kpt pkg get "$SDK_REPO"/example-configs example-configs
kpt pkg get https://github.com/prachirp/kpt-functions-catalog/functions/go@e2e-revamp ./
kpt fn run example-configs --enable-exec --exec-path "$(pwd)"/go/set-namespace/set-namespace-linux -- namespace=example-ns
assert_contains_string example-configs/gatekeeper.yaml "namespace: example-ns"

testcase "kpt_set_namespace_go_docker_imperative_success"
kpt pkg get "$SDK_REPO"/example-configs example-configs
kpt fn run . --image gcr.io/kpt-functions/set-namespace -- namespace=example-ns
assert_contains_string example-configs/gatekeeper.yaml "namespace: example-ns"

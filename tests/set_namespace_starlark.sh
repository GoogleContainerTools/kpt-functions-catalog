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
# E2E tests for set_namespace_starlark.

set -eo pipefail
DIR="$(dirname "$0")"
# shellcheck source=tests/common.sh
source "$DIR"/common.sh

# TODO: Write imperative e2e test calling starlark
# Blocked because passing arguments to `kpt fn run` puts them in the data field instead of the spec field
# https://github.com/GoogleContainerTools/kpt/issues/757
testcase "kpt_set_namespace_starlark_declarative"
kpt pkg get "$SDK_REPO"/example-configs example-configs
kpt pkg get "$CATALOG_REPO"/functions/starlark ./
cat >fc.yaml <<EOF
apiVersion: example.com/v1beta1
kind: ExampleKind
metadata:
  name: function-input
  namespace: example-ns
  annotations:
    config.kubernetes.io/function: |
      starlark: {path: starlark/set_namespace.star, name: example-name}
spec:
  namespace_value: example-ns
EOF
kpt fn run . --enable-star
assert_contains_string example-configs/gatekeeper.yaml "namespace: example-ns"

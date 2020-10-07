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
# E2E tests for sops.

set -eo pipefail
DIR="$(dirname "$0")"
# shellcheck source=tests/common.sh
source "$DIR"/common.sh

testcase "kpt_sops_imperative_expected_args"
mkdir example-configs && curl -fsSL -o example-configs/example.yaml https://raw.githubusercontent.com/mozilla/sops/master/example.yaml || echo "couldn't create example.yaml"
curl -fsSL -o key.asc https://raw.githubusercontent.com/mozilla/sops/master/pgp/sops_functional_tests_key.asc || echo "couldn't create key.asc"
kpt fn source example-configs |
  kpt fn run --env SOPS_IMPORT_PGP="$(cat key.asc)" --image gcr.io/kpt-functions/sops:"${TAG}" -- verbose=true >out.yaml
assert_contains_string out.yaml "t00m4nys3cr3tzupdated"

if false; then
# TODO: remove 'if' when gcr.io/kpt-functions/sops:latest is available
testcase "kpt_sops_declarative_example"
# TODO: Remove error handling once kpt pkg get shows errors gracefully https://github.com/GoogleContainerTools/kpt/issues/838
# TODO: This is a temporary fix till declarative example e2e tests run using the $TAG being tested
kpt pkg get "$CATALOG_REPO"/examples/sops . || true
curl -fsSL -o key.asc https://raw.githubusercontent.com/mozilla/sops/master/pgp/sops_functional_tests_key.asc
kpt fn run --env SOPS_IMPORT_PGP="$(cat key.asc)" sops/local-configs
assert_contains_string sops/local-configs/encrypted-gvk.yaml "nnn-password: k8spassphrase"
fi

testcase "kpt_sops_declarative_fn_path"
cat >fc.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  annotations:
    config.k8s.io/function: |
      container:
        image: gcr.io/kpt-functions/sops:${TAG}
    config.kubernetes.io/local-config: "true"
data:
EOF
mkdir example-configs && curl -fsSL -o example-configs/example.yaml https://raw.githubusercontent.com/mozilla/sops/master/example.yaml || echo "couldn't create example.yaml"
curl -fsSL -o key.asc https://raw.githubusercontent.com/mozilla/sops/master/pgp/sops_functional_tests_key.asc || echo "couldn't create key.asc"
kpt fn source example-configs |
  kpt fn run --env SOPS_IMPORT_PGP="$(cat key.asc)" --fn-path fc.yaml >out.yaml
assert_contains_string out.yaml "t00m4nys3cr3tzupdated"

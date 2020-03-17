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
# E2E tests for kpt-functions-catalog.

set -eo pipefail

TAG=${TAG:-dev}
SDK_REPO=https://github.com/GoogleContainerTools/kpt-functions-sdk
HELM_USAGE_SNIPPET="Render chart templates locally"
CHARTS_SRC="charts/bitnami"

############################
# Test framework
############################

function testcase() {
  echo "testcase: ${1}"
  tmp=$(mktemp -d "/tmp/e2e.${1}.XXXXXXXX")
  cd "${tmp}"
  kpt pkg get $SDK_REPO/example-configs example-configs
  git clone -q https://github.com/bitnami/charts.git
}

function fail() {
  echo "FAIL: " "$@"
  exit 1
}

function assert_contains_string() {
  content="$(<"$1")"
  grep -q "$2" "$1" || fail "String $2 not contained in: ${content}"
}

function assert_dir_exists() {
  [[ -d $1 ]] || fail "Dir not exist: $1"
}

############################
# Docker Tests
############################

testcase "docker_helm_template_usage_too_few_args"
docker run -v "$(pwd)/${CHARTS_SRC}":/source gcr.io/kpt-functions/helm-template:"${TAG}" name=too-few-args 2>err.txt || true
assert_contains_string err.txt "${HELM_USAGE_SNIPPET}"

testcase "docker_helm_template_usage_too_many_args"
docker run -v "$(pwd)/${CHARTS_SRC}":/source gcr.io/kpt-functions/helm-template:"${TAG}" chart_path=/source/redis name=too-many-args extra=args_provided 2>err.txt || true
assert_contains_string err.txt "${HELM_USAGE_SNIPPET}"

testcase "docker_helm_template_correct_args"
docker run -v "$(pwd)/${CHARTS_SRC}":/source gcr.io/kpt-functions/helm-template:"${TAG}" name=correct-args chart_path=/source/redis >out.yaml
assert_contains_string out.yaml "correct-args"

testcase "docker_helm_template_values_arg_sink"
docker run -v "$(pwd)/${CHARTS_SRC}":/source gcr.io/kpt-functions/helm-template:"${TAG}" chart_path=/source/redis name=prod-values values_path=/source/redis/values-production.yaml |
  docker run -i -u "$(id -u)" -v "$(pwd)":/sink gcr.io/kpt-functions/write-yaml:"${TAG}" -o /dev/null -d sink_dir=/sink -d overwrite=true
assert_dir_exists redis

testcase "docker_helm_template_stdin_correct_args_sink"
docker run -v "$(pwd)/${CHARTS_SRC}":/source gcr.io/kpt-functions/helm-template chart_path=/source/mongodb name=my-mongodb |
  docker run -i -v "$(pwd)/${CHARTS_SRC}":/source gcr.io/kpt-functions/helm-template name=my-redis chart_path=/source/redis |
  docker run -i -u "$(id -u)" -v "$(pwd)":/sink gcr.io/kpt-functions/write-yaml:"${TAG}" -o /dev/null -d sink_dir=/sink -d overwrite=true
assert_dir_exists mongodb
assert_dir_exists redis

############################
# kpt fn Tests
############################

# TODO: Add kpt_helm_template_short test after kpt fn run --image fix <https://github.com/GoogleContainerTools/kpt/issues/359>
testcase "kpt_helm_template_imperative"
kpt fn source example-configs |
  docker run -i -v "$(pwd)/${CHARTS_SRC}":/source gcr.io/kpt-functions/helm-template chart_path=/source/mongodb name=my-mongodb |
  docker run -i -v "$(pwd)/${CHARTS_SRC}":/source gcr.io/kpt-functions/helm-template name=my-redis chart_path=/source/redis |
  kpt fn sink example-configs
assert_dir_exists ./example-configs/mongodb
assert_dir_exists ./example-configs/redis

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
HELM_ERROR_SNIPPET="Helm template command results in error"
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

testcase "docker_helm_template_undefined_args"
docker run -u "$(id -u)" -v "$(pwd)/${CHARTS_SRC}":/source gcr.io/kpt-functions/helm-template:"${TAG}" -i /dev/null 2>err.txt || true
assert_contains_string err.txt "Error: functionConfig expected, instead undefined"

testcase "docker_helm_template_empty_fc"
cat >fc.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: empty-config
  annotations:
    config.k8s.io/function: |
      container:
        image:  gcr.io/kpt-functions/helm-template
    config.kubernetes.io/local-config: "true"
data:
EOF
docker run -u "$(id -u)" -v "$(pwd)":/source gcr.io/kpt-functions/helm-template:"${TAG}" -i /dev/null -f /source/fc.yaml 2>err.txt || true
assert_contains_string err.txt "Error: functionConfig expected to contain data, instead empty"

testcase "docker_helm_template_invalid_fc"
cat >fc.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: invalid-config
  annotations:
    config.k8s.io/function: |
      container:
        image:  gcr.io/kpt-functions/helm-template
    config.kubernetes.io/local-config: "true"
data:
  name: invalid-fc
  chart_path: /path/to/helm/chart
EOF
docker run -u "$(id -u)" -v "$(pwd)":/source gcr.io/kpt-functions/helm-template:"${TAG}" -i /dev/null -f /source/fc.yaml >out.yaml || true
assert_contains_string out.yaml "${HELM_ERROR_SNIPPET}"

testcase "docker_helm_template_too_few_args"
docker run -u "$(id -u)" -v "$(pwd)/${CHARTS_SRC}":/source gcr.io/kpt-functions/helm-template:"${TAG}" -i /dev/null -d name=too-few-args >out.yaml || true
assert_contains_string out.yaml "${HELM_ERROR_SNIPPET}"

testcase "docker_helm_template_expected_args"
docker run -u "$(id -u)" -v "$(pwd)/${CHARTS_SRC}":/source gcr.io/kpt-functions/helm-template:"${TAG}" -i /dev/null -d name=expected-args -d chart_path=/source/redis >out.yaml
assert_contains_string out.yaml "expected-args"

testcase "docker_helm_template_extra_args"
cat >fc.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  annotations:
    config.k8s.io/function: |
      container:
        image:  gcr.io/kpt-functions/helm-template
    config.kubernetes.io/local-config: "true"
data:
  name: extra-args
  chart_path: /source/charts/bitnami/redis
  --values: /source/charts/bitnami/redis/values-production.yaml
EOF
docker run -u "$(id -u)" -v "$(pwd)":/source gcr.io/kpt-functions/helm-template:"${TAG}" -i /dev/null -f /source/fc.yaml >out.yaml
assert_contains_string out.yaml "extra-args"

testcase "docker_helm_template_sink"
docker run -u "$(id -u)" -v "$(pwd)/${CHARTS_SRC}":/source gcr.io/kpt-functions/helm-template:"${TAG}" -i /dev/null -d chart_path=/source/redis -d name=sink-redis |
  docker run -i -u "$(id -u)" -v "$(pwd)":/sink gcr.io/kpt-functions/write-yaml:"${TAG}" -o /dev/null -d sink_dir=/sink -d overwrite=true
assert_dir_exists default
assert_contains_string default/secret_sink-redis.yaml "sink-redis"

testcase "docker_helm_template_pipeline"
docker run -u "$(id -u)" -v "$(pwd)/${CHARTS_SRC}":/source gcr.io/kpt-functions/helm-template:"${TAG}" -i /dev/null -d chart_path=/source/mongodb -d name=my-mongodb |
  docker run -i -u "$(id -u)" -v "$(pwd)/${CHARTS_SRC}":/source gcr.io/kpt-functions/helm-template:"${TAG}" -d name=my-redis -d chart_path=/source/redis |
  docker run -i -u "$(id -u)" -v "$(pwd)":/sink gcr.io/kpt-functions/write-yaml:"${TAG}" -o /dev/null -d sink_dir=/sink -d overwrite=true
assert_dir_exists default
assert_contains_string default/secret_my-mongodb.yaml "my-mongodb"
assert_contains_string default/secret_my-redis.yaml "my-redis"

############################
# kpt fn Tests
############################

# TODO: Add kpt_helm_template_imperative_short and kpt_helm_template_declarative tests after fixing <https://github.com/GoogleContainerTools/kpt/issues/587>
testcase "kpt_helm_template_imperative"
kpt fn source example-configs |
  docker run -i -u "$(id -u)" -v "$(pwd)/${CHARTS_SRC}":/source gcr.io/kpt-functions/helm-template:"${TAG}" -d chart_path=/source/mongodb -d name=my-mongodb |
  docker run -i -u "$(id -u)" -v "$(pwd)/${CHARTS_SRC}":/source gcr.io/kpt-functions/helm-template:"${TAG}" -d name=my-redis -d chart_path=/source/redis |
  kpt fn sink example-configs
assert_dir_exists example-configs/default
assert_contains_string example-configs/default/secret_my-mongodb.yaml "my-mongodb"
assert_contains_string example-configs/default/secret_my-redis.yaml "my-redis"

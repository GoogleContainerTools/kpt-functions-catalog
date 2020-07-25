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
# E2E tests for helm_template.

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

# TODO: Convert source function tests to kpt fn after fixing https://github.com/GoogleContainerTools/kpt/issues/587
helm_testcase "docker_helm_template_imperative_expected_args"
docker run -u "$(id -u)" -v "$(pwd)/${CHARTS_SRC}":/source gcr.io/kpt-functions/helm-template:"${TAG}" -i /dev/null -d name=expected-args -d chart_path=/source/redis >out.yaml
assert_contains_string out.yaml "expected-args"

helm_testcase "docker_helm_template_declarative_extra_args"
cat >fc.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  annotations:
    config.k8s.io/function: |
      container:
        image:  gcr.io/kpt-functions/helm-template:${TAG}
    config.kubernetes.io/local-config: "true"
data:
  name: extra-args
  chart_path: /source/charts/bitnami/redis
  --values: /source/charts/bitnami/redis/values-production.yaml
EOF
docker run -u "$(id -u)" -v "$(pwd)":/source gcr.io/kpt-functions/helm-template:"${TAG}" -i /dev/null -f /source/fc.yaml >out.yaml
assert_contains_string out.yaml "name: extra-args-redis-master"

helm_testcase "docker_helm_template_imperative_sink"
docker run -u "$(id -u)" -v "$(pwd)/${CHARTS_SRC}":/source gcr.io/kpt-functions/helm-template:"${TAG}" -i /dev/null -d chart_path=/source/redis -d name=sink-redis |
  docker run -i -u "$(id -u)" -v "$(pwd)":/sink gcr.io/kpt-functions/write-yaml:"${TAG}" -o /dev/null -d sink_dir=/sink -d overwrite=true
assert_dir_exists default
assert_contains_string default/secret_sink-redis.yaml "name: sink-redis"

helm_testcase "docker_helm_template_imperative_pipeline"
git clone -q https://github.com/helm/charts.git helm-charts
docker run -u "$(id -u)" -v "$(pwd)/helm-charts":/source gcr.io/kpt-functions/helm-template:"${TAG}" -i /dev/null -d chart_path=/source/incubator/haproxy-ingress -d name=my-haproxy-ingress | 
  docker run -i -u "$(id -u)" -v "$(pwd)/${CHARTS_SRC}":/source gcr.io/kpt-functions/helm-template:"${TAG}" -d name=my-redis -d chart_path=/source/redis |
  docker run -i -u "$(id -u)" -v "$(pwd)":/sink gcr.io/kpt-functions/write-yaml:"${TAG}" -o /dev/null -d sink_dir=/sink -d overwrite=true
assert_dir_exists default
assert_contains_string default/configmap_my-haproxy-ingress-controller.yaml "name: my-haproxy-ingress"
assert_contains_string default/secret_my-redis.yaml "name: my-redis"

############################
# kpt fn Tests
############################

helm_testcase "kpt_helm_template_imperative_expected_args"
kpt fn source example-configs |
  kpt fn run --mount type=bind,src="$(pwd)/${CHARTS_SRC}",dst=/source --image gcr.io/kpt-functions/helm-template:"${TAG}" -- name=expected-args chart_path=/source/redis >out.yaml
assert_contains_string out.yaml "expected-args"

testcase "kpt_helm_template_declarative_example"
# TODO: Remove error handling once kpt pkg get shows errors gracefully https://github.com/GoogleContainerTools/kpt/issues/838
kpt pkg get "$CATALOG_REPO"/examples/helm-template . || true
kpt fn run helm-template/local-configs --mount type=bind,src="$(pwd)"/helm-template/helloworld-chart,dst=/source
assert_contains_string helm-template/local-configs/deployment_chart-helloworld-chart.yaml "name: chart-helloworld-chart"

helm_testcase "kpt_helm_template_declarative_fn_path"
cat >fc.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  annotations:
    config.k8s.io/function: |
      container:
        image:  gcr.io/kpt-functions/helm-template:${TAG}
    config.kubernetes.io/local-config: "true"
data:
  name: extra-args
  chart_path: /source
  --values: /source/values-production.yaml
EOF
kpt fn source example-configs |
  kpt fn run --mount type=bind,src="$(pwd)/${CHARTS_SRC}/redis",dst=/source --fn-path fc.yaml >out.yaml
assert_contains_string out.yaml "name: extra-args-redis-master"

helm_testcase "kpt_helm_template_imperative_pipeline"
git clone -q https://github.com/helm/charts.git helm-charts
kpt fn source example-configs |
  kpt fn run --mount type=bind,src="$(pwd)/helm-charts",dst=/source --image gcr.io/kpt-functions/helm-template:"${TAG}" -- chart_path=/source/incubator/haproxy-ingress name=my-haproxy-ingress | 
  kpt fn run --mount type=bind,src="$(pwd)/${CHARTS_SRC}",dst=/source --image gcr.io/kpt-functions/helm-template:"${TAG}" -- name=my-redis chart_path=/source/redis |
  kpt fn sink .
assert_dir_exists default
assert_contains_string default/configmap_my-haproxy-ingress-controller.yaml "name: my-haproxy-ingress"
assert_contains_string default/secret_my-redis.yaml "name: my-redis"

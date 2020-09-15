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

helm_testcase "kpt_helm_template_imperative_expected_args"
kpt fn source example-configs |
  kpt fn run --mount type=bind,src="$(pwd)/${CHARTS_SRC}",dst=/source --image gcr.io/kpt-functions/helm-template:"${TAG}" -- name=expected-args chart_path=/source/redis >out.yaml
assert_contains_string out.yaml "expected-args"

testcase "kpt_helm_template_declarative_example"
# TODO: Remove error handling once kpt pkg get shows errors gracefully https://github.com/GoogleContainerTools/kpt/issues/838
# TODO: This is a temporary fix till declarative example e2e tests run using the $TAG being tested
kpt pkg get https://github.com/prachirp/kpt-functions-catalog.git/examples/helm-template@helm-fix . || true
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
        image: gcr.io/kpt-functions/helm-template:${TAG}
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

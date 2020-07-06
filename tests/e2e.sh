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
# E2E tests for ts-functions.

set -eo pipefail

TAG=${TAG:-dev}
NODOCKER=${NODOCKER:-}
SDK_REPO=https://github.com/GoogleContainerTools/kpt-functions-sdk
CATALOG_REPO=https://github.com/GoogleContainerTools/kpt-functions-catalog
CHARTS_SRC="charts/bitnami"

############################
# Test framework
############################

function testcase() {
  echo "testcase: ${1}"
  tmp=$(mktemp -d "/tmp/e2e.${1}.XXXXXXXX")
  cd "${tmp}"
}

function helm_testcase() {
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
[[ -z ${NODOCKER} ]] || {
  echo "Skipping docker tests"
  exit 0
}

# TODO: Convert source function tests to kpt fn after fixing https://github.com/GoogleContainerTools/kpt/issues/587
helm_testcase "docker_helm_template_expected_args"
docker run -u "$(id -u)" -v "$(pwd)/${CHARTS_SRC}":/source gcr.io/kpt-functions/helm-template:"${TAG}" -i /dev/null -d name=expected-args -d chart_path=/source/redis >out.yaml
assert_contains_string out.yaml "expected-args"

helm_testcase "docker_helm_template_extra_args"
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

helm_testcase "docker_helm_template_sink"
docker run -u "$(id -u)" -v "$(pwd)/${CHARTS_SRC}":/source gcr.io/kpt-functions/helm-template:"${TAG}" -i /dev/null -d chart_path=/source/redis -d name=sink-redis |
  docker run -i -u "$(id -u)" -v "$(pwd)":/sink gcr.io/kpt-functions/write-yaml:"${TAG}" -o /dev/null -d sink_dir=/sink -d overwrite=true
assert_dir_exists default
assert_contains_string default/secret_sink-redis.yaml "name: sink-redis"

helm_testcase "docker_helm_template_pipeline"
git clone -q https://github.com/helm/charts.git helm-charts
docker run -u "$(id -u)" -v "$(pwd)/helm-charts":/source gcr.io/kpt-functions/helm-template:"${TAG}" -i /dev/null -d chart_path=/source/incubator/haproxy-ingress -d name=my-haproxy-ingress | 
  docker run -i -u "$(id -u)" -v "$(pwd)/${CHARTS_SRC}":/source gcr.io/kpt-functions/helm-template:"${TAG}" -d name=my-redis -d chart_path=/source/redis |
  docker run -i -u "$(id -u)" -v "$(pwd)":/sink gcr.io/kpt-functions/write-yaml:"${TAG}" -o /dev/null -d sink_dir=/sink -d overwrite=true
assert_dir_exists default
assert_contains_string default/configmap_my-haproxy-ingress-controller.yaml "name: my-haproxy-ingress"
assert_contains_string default/secret_my-redis.yaml "name: my-redis"

testcase "docker_kustomize_build_git"
docker run -u "$(id -u)" gcr.io/kpt-functions/kustomize-build -d path=https://github.com/kubernetes-sigs/kustomize/examples/multibases/ |
  docker run -i -u "$(id -u)" -v "$(pwd)":/sink gcr.io/kpt-functions/write-yaml:"${TAG}" -o /dev/null -d sink_dir=/sink -d overwrite=true
assert_contains_string pod_cluster-a-staging-myapp-pod.yaml "name: cluster-a-staging-myapp-pod"

testcase "docker_kustomize_build_extra_args"
kpt pkg get https://github.com/kubernetes-sigs/kustomize/examples/helloWorld helloWorld
cat >fc.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  annotations:
    config.k8s.io/function: |
      container:
        image:  gcr.io/kpt-functions/kustomize-build:${TAG}
    config.kubernetes.io/local-config: "true"
data:
  path: /source/helloWorld
  --output: /source/kustomize_build_output.yaml
EOF
docker run -u "$(id -u)" --mount type=bind,src="$(pwd)",dst=/source gcr.io/kpt-functions/kustomize-build:"${TAG}" -f /source/fc.yaml
assert_contains_string kustomize_build_output.yaml "app: hello"

testcase "docker_kustomize_build_sink"
kpt pkg get https://github.com/kubernetes-sigs/kustomize/examples/helloWorld helloWorld
docker run -u "$(id -u)" --mount type=bind,src="$(pwd)",dst=/source gcr.io/kpt-functions/kustomize-build:"${TAG}" -d path=/source/helloWorld |
  docker run -i -u "$(id -u)" -v "$(pwd)":/sink gcr.io/kpt-functions/write-yaml:"${TAG}" -o /dev/null -d sink_dir=/sink -d overwrite=true
assert_contains_string configmap_the-map.yaml "app: hello"

testcase "docker_kustomize_build_pipeline"
kpt pkg get https://github.com/kubernetes-sigs/kustomize/examples examples
docker run -u "$(id -u)" --mount type=bind,src="$(pwd)/examples",dst=/source gcr.io/kpt-functions/kustomize-build:"${TAG}" -d path=/source/loadHttp |
  docker run -i -u "$(id -u)" --mount type=bind,src="$(pwd)/examples",dst=/source gcr.io/kpt-functions/kustomize-build:"${TAG}" -d path=/source/helloWorld |
  docker run -i -u "$(id -u)" -v "$(pwd)":/sink gcr.io/kpt-functions/write-yaml:"${TAG}" -o /dev/null -d sink_dir=/sink -d overwrite=true
assert_contains_string configmap_the-map.yaml "app: hello"
assert_dir_exists knative-serving

############################
# kpt fn Tests
############################

testcase "kpt_set_namespace_success"
kpt pkg get $SDK_REPO/example-configs example-configs
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
kpt pkg get $CATALOG_REPO/functions/starlark ./
kpt fn run . --enable-star
assert_contains_string example-configs/gatekeeper.yaml "namespace: example-ns"

testcase "kpt_istioctl_analyze_success"
kpt pkg get https://github.com/istio/istio.git/samples/addons .
cat >fc.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  annotations:
    config.k8s.io/function: |
      container:
        image:  gcr.io/kpt-functions/istioctl-analyze:${TAG}
    config.kubernetes.io/local-config: 'true'
data:
  "flags": [ "--recursive" ]
  "--use-kube": "false"
EOF
kpt fn source addons | kpt fn run --fn-path fc.yaml 2>error.txt | kpt fn sink addons
if [ -s error.txt ]; then
  fail "Validation error found using istio addons sample: " "$(< error.txt)"
fi

testcase "kpt_istioctl_analyze_error"
kpt pkg get https://github.com/istio/istio.git/galley/pkg/config/analysis/analyzers/testdata .
cat >fc.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  annotations:
    config.k8s.io/function: |
      container:
        image:  gcr.io/kpt-functions/istioctl-analyze:${TAG}
    config.kubernetes.io/local-config: 'true'
data:
  "flags": [ "--recursive" ]
  "--use-kube": "false"
EOF
kpt fn run testdata --fn-path fc.yaml 2>error.txt || true
assert_contains_string error.txt "Referenced selector not found"

testcase "kpt_kubeval_success"
kpt pkg get https://github.com/instrumenta/kubeval.git/fixtures .
cat >fc.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  annotations:
    config.k8s.io/function: |
      container:
        image:  gcr.io/kpt-functions/kubeval:${TAG}
        network:
          required: true
    config.kubernetes.io/local-config: 'true'
EOF
kpt fn source fixtures/valid* |
kpt fn run --fn-path fc.yaml --network >out.yaml
if grep -q "results" out.yaml; then
  fail "Validation error found using kubeval fixtures valid config: " "$(< out.yaml)"
fi

testcase "kpt_kubeval_error"
kpt pkg get https://github.com/instrumenta/kubeval.git/fixtures .
cat >fc.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  annotations:
    config.k8s.io/function: |
      container:
        image:  gcr.io/kpt-functions/kubeval:${TAG}
        network:
          required: true
    config.kubernetes.io/local-config: 'true'
EOF
kpt fn source fixtures/*invalid.yaml |
kpt fn run --fn-path fc.yaml --network 2>error.txt || true
assert_contains_string error.txt "Invalid type. Expected: \[integer,null\], given: string"

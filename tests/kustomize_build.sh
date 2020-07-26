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
# E2E tests for kustomize_build.

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

testcase "docker_kustomize_build_imperative_git"
docker run -u "$(id -u)" gcr.io/kpt-functions/kustomize-build:"${TAG}" -d path=https://github.com/kubernetes-sigs/kustomize/examples/multibases/ |
  docker run -i -u "$(id -u)" -v "$(pwd)":/sink gcr.io/kpt-functions/write-yaml:"${TAG}" -o /dev/null -d sink_dir=/sink -d overwrite=true
assert_contains_string pod_cluster-a-staging-myapp-pod.yaml "name: cluster-a-staging-myapp-pod"

testcase "docker_kustomize_build_declarative_extra_args"
kpt pkg get https://github.com/kubernetes-sigs/kustomize/examples/helloWorld helloWorld
cat >fc.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  annotations:
    config.k8s.io/function: |
      container:
        image: gcr.io/kpt-functions/kustomize-build:${TAG}
    config.kubernetes.io/local-config: "true"
data:
  path: /source/helloWorld
  --output: /source/kustomize_build_output.yaml
EOF
docker run -u "$(id -u)" --mount type=bind,src="$(pwd)",dst=/source gcr.io/kpt-functions/kustomize-build:"${TAG}" -f /source/fc.yaml
assert_contains_string kustomize_build_output.yaml "app: hello"

testcase "docker_kustomize_build_imperative_sink"
kpt pkg get https://github.com/kubernetes-sigs/kustomize/examples/helloWorld helloWorld
docker run -u "$(id -u)" --mount type=bind,src="$(pwd)",dst=/source gcr.io/kpt-functions/kustomize-build:"${TAG}" -d path=/source/helloWorld |
  docker run -i -u "$(id -u)" -v "$(pwd)":/sink gcr.io/kpt-functions/write-yaml:"${TAG}" -o /dev/null -d sink_dir=/sink -d overwrite=true
assert_contains_string configmap_the-map.yaml "app: hello"

testcase "docker_kustomize_build_imperative_pipeline"
kpt pkg get https://github.com/kubernetes-sigs/kustomize/examples examples
docker run -u "$(id -u)" --mount type=bind,src="$(pwd)/examples",dst=/source gcr.io/kpt-functions/kustomize-build:"${TAG}" -d path=/source/loadHttp |
  docker run -i -u "$(id -u)" --mount type=bind,src="$(pwd)/examples",dst=/source gcr.io/kpt-functions/kustomize-build:"${TAG}" -d path=/source/helloWorld |
  docker run -i -u "$(id -u)" -v "$(pwd)":/sink gcr.io/kpt-functions/write-yaml:"${TAG}" -o /dev/null -d sink_dir=/sink -d overwrite=true
assert_contains_string configmap_the-map.yaml "app: hello"
assert_dir_exists knative-serving

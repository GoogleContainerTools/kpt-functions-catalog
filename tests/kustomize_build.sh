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

testcase "kpt_kustomize_build_imperative"
kpt pkg get https://github.com/kubernetes-sigs/kustomize/examples/helloWorld helloWorld
kpt fn source helloWorld |
  kpt fn run --mount type=bind,src="$(pwd)/helloWorld",dst=/source --image gcr.io/kpt-functions/kustomize-build:"${TAG}" -- path=/source |
  kpt fn sink .
assert_contains_string configmap_the-map.yaml "app: hello"

testcase "kpt_kustomize_build_declarative_example"
# TODO: Remove error handling once kpt pkg get shows errors gracefully https://github.com/GoogleContainerTools/kpt/issues/838
kpt pkg get "$CATALOG_REPO"/examples/kustomize-build . || true
kpt fn run kustomize-build/local-configs --mount type=bind,src="$(pwd)"/kustomize-build/kustomize-dir,dst=/source
assert_contains_string kustomize-build/local-configs/configmap_example-cm.yaml "name: example-cm"

testcase "kpt_kustomize_build_declarative_fn_path"
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
  path: /source
  --load_restrictor: LoadRestrictionsNone
EOF
kpt fn run . --mount type=bind,src="$(pwd)/helloWorld",dst=/source,rw=true --fn-path fc.yaml
assert_contains_string configmap_the-map.yaml "app: hello"

testcase "kpt_kustomize_build_imperative_pipeline"
kpt pkg get https://github.com/kubernetes-sigs/kustomize/examples examples
kpt fn source examples |
  kpt fn run --mount type=bind,src="$(pwd)/examples/loadHttp",dst=/source --network --image gcr.io/kpt-functions/kustomize-build:"${TAG}" -- path=/source |
  kpt fn run --mount type=bind,src="$(pwd)/examples/helloWorld",dst=/source --network --image gcr.io/kpt-functions/kustomize-build:"${TAG}" -- path=/source |
  kpt fn sink .
assert_contains_string configmap_the-map.yaml "app: hello"
assert_dir_exists knative-serving

testcase "kpt_kustomize_build_declarative_git"
cat >fc.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  annotations:
    config.k8s.io/function: |
      container:
        image: gcr.io/kpt-functions/kustomize-build:${TAG}
        network:
          required: true
    config.kubernetes.io/local-config: "true"
data:
  path: http://github.com/kubernetes-sigs/kustomize/examples/multibases/
EOF
kpt fn run . --network
assert_contains_string pod_cluster-a-staging-myapp-pod.yaml "name: cluster-a-staging-myapp-pod"

testcase "kpt_kustomize_build_imperative_git"
kpt fn run . --network --image gcr.io/kpt-functions/kustomize-build:"${TAG}" -- path=http://github.com/kubernetes-sigs/kustomize/examples/multibases/
assert_contains_string pod_cluster-a-staging-myapp-pod.yaml "name: cluster-a-staging-myapp-pod"

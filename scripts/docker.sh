#! /bin/bash
#
# Copyright 2021 Google LLC
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

repo_base=$(cd "$(dirname "$(dirname "$0")")" || exit ; pwd)

UNSTABLE_TAG=unstable
DEFAULT_GCR=${DEFAULT_GCR:-gcr.io/kpt-fn-contrib}
GCR_REGISTRY=${GCR_REGISTRY:-${DEFAULT_GCR}}

function err {
  echo "$1"
  exit 1
}

function docker_build {
  type=$1 # function type, e.g. contrib, curated
  lang=$2 # function language, e.g. go, ts
  name=$3 # function name, e.g. apply-setters

  build_args=()

  case "${type}" in
    contrib) function_dir="${repo_base}/contrib/functions/${lang}/${name}" ;;
    curated) function_dir="${repo_base}/functions/${lang}/${name}" ;;
    *) err "unknown function type: ${type}" ;;
  esac

  case "${lang}" in
    ts)
      translated_name=$(echo "${name}" | tr - _)
      build_args+=(--build-arg "FILENAME=${translated_name}_run.js")
      override_dockerfile="${function_dir}/build/${translated_name}.Dockerfile"
      ;;
    *) override_dockerfile="${function_dir}"/Dockerfile ;;
  esac

  dockerfile="${repo_base}/build/docker/${lang}/Dockerfile"
  [[ -f "${override_dockerfile}" ]] && dockerfile="${override_dockerfile}"
  [[ -f "${dockerfile}" ]] || err "Dockerfile does not exist: ${dockerfile}"

  # Use + conditional parameter expansion to protect from unbound array variable
  docker buildx build \
    --platform=linux/amd64,linux/arm64 \
    -t "${GCR_REGISTRY}/${name}:${UNSTABLE_TAG}" \
    -f "${dockerfile}" \
    "${build_args[@]+"${build_args[@]}"}" \
    "${function_dir}"
}

function docker_push {
  name=$1 # function name, e.g. apply-setters
  version=$2 # function version, e.g. v0.1.1

  docker push "${GCR_REGISTRY}/${name}:${version}"
}


function docker_tag {
  name=$1 # function name, e.g. apply-setters
  version=$2 # function version, e.g. v0.1.1

  echo tagging "${GCR_REGISTRY}/${name}:${version}"
  docker tag "${GCR_REGISTRY}/${name}:${UNSTABLE_TAG}" "${GCR_REGISTRY}/${name}:${version}"
}

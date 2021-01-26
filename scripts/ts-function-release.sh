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

set -euo pipefail

scripts_dir="$(dirname "$0")"
# git-tag-parser.sh has been shell-checked separately.
# shellcheck source=/dev/null
source ${scripts_dir}/git-tag-parser.sh

DEV_TAG=dev
versions=$(get_versions "${TAG}")
GCR_REGISTRY=${GCR_REGISTRY:-gcr.io/kpt-functions}

cd "${scripts_dir}/../functions/ts/${CURRENT_FUNCTION}"

# This make it work for npm 6.*.*
sed -i.bak "s|gcr.io/kpt-functions|${GCR_REGISTRY}|g" package.json

# This make it work for npm 7.0.0+
export npm_package_kpt_docker_repo_base="${GCR_REGISTRY}"

case "$1" in
  build)
    npm run kpt:docker-build -- --tag ${DEV_TAG}
    for version in ${versions}; do
      echo tagging "${GCR_REGISTRY}/${CURRENT_FUNCTION}:${version}"
      docker tag "${GCR_REGISTRY}/${CURRENT_FUNCTION}:${DEV_TAG}" "${GCR_REGISTRY}/${CURRENT_FUNCTION}:${version}"
    done
    ;;
  push)
    for version in ${versions}; do
      docker tag "${GCR_REGISTRY}/${CURRENT_FUNCTION}:${DEV_TAG}" "${GCR_REGISTRY}/${CURRENT_FUNCTION}:${version}"
      npm run kpt:docker-push -- --tag="${version}"
    done
    ;;
  *)
    echo "Usage: $0 {build|push}"
    exit 1
esac

# Change it back
sed -i.bak "s|${GCR_REGISTRY}|gcr.io/kpt-functions|g" package.json

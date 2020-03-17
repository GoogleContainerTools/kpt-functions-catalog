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
# Publish shell functions to catalog.

set -eo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. >/dev/null 2>&1 && pwd)"
TAG=${TAG:-dev}
FUNCS=(helm-template)

for FN_NAME in "${FUNCS[@]}"; do
  DOCKER_IMAGE="gcr.io/kpt-functions/${FN_NAME}:${TAG}"
  DOCKER_FILE="${REPO}/functions/${FN_NAME}/Dockerfile"
  docker build -t "$DOCKER_IMAGE" -f "$DOCKER_FILE" functions
  docker push "$DOCKER_IMAGE"
done

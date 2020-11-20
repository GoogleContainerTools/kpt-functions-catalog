#!/bin/bash
# Copyright 2020 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

# if `BUILDONLY` env var isn't defined set it empty
BUILDONLY=${BUILDONLY:-}

# if the `TAG` environment variable is not defined, set image_tag to `latest`.
image_tag=${TAG:-latest}

# if `image_tag` starts with "release-go-functions-", remove the prefix.
prefix="release-go-functions-"
[[ "${image_tag}" = "${prefix}"* ]] && image_tag="${image_tag#$prefix}"

make

# iterate over each subdir, build and push Docker images.
for dir in  */
do
  image_name=gcr.io/kpt-functions/"${dir%/}"
  image="${image_name}":"${image_tag}"
  set -x
  docker build -t "${image}" -t "${image_name}" -f "${dir}"/Dockerfile "${dir}"
  if [ -z "${BUILDONLY}" ]; then
    docker push "${image_name}"
  fi
  set +x
done

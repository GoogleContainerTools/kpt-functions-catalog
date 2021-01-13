#! /bin/bash

set -euo pipefail

source ../../scripts/get-versions.sh

image_tag=${TAG:-dev}
tags=$(validate_version ${image_tag})
prefix="release-go-functions-"
[[ "${image_tag}" = "${prefix}"* ]] && image_tag="${image_tag#$prefix}"
image_name=gcr.io/kpt-fn/"${CURRENT_FUNCTION}"
tag_params=""
for tag in ${tags}; do
  tag_params="$tag_params -t ${image_name}:${tag}"
done

docker build ${tag_params} -f "${CURRENT_FUNCTION}"/Dockerfile "${CURRENT_FUNCTION}"

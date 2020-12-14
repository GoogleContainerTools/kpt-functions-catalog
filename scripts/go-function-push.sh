#! /bin/bash

set -euo pipefail

image_tag=${TAG:-dev}
prefix="release-go-functions-"
[[ "${image_tag}" = "${prefix}"* ]] && image_tag="${image_tag#$prefix}"
image_name=gcr.io/kpt-functions/"${CURRENT_FUNCTION}"
image="${image_name}":"${image_tag}"
docker push "${image_name}:unstable"
docker push "${image}"

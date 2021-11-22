#!/usr/bin/env bash

set -eo pipefail

kpt fn eval -i gcr.io/kpt-fn/kubeval:unstable --image-pull-policy never \
  --mount type=bind,src="$(pwd)/jsonschema",dst=/schema-dir/master-standalone \
  -- schema_location=file:///schema-dir

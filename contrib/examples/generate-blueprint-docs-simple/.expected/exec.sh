#!/usr/bin/env bash

set -eo pipefail

kpt fn eval -i gcr.io/kpt-fn-contrib/generate-blueprint-docs:unstable --image-pull-policy never \
--include-meta-resources --mount type=bind,src="$(pwd)",dst=/tmp,rw=true -- readme-path=/tmp/GENERATED.md
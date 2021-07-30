#!/usr/bin/env bash

docker run --rm -v "$(pwd)"/resources:/resources gcr.io/kpt-fn/gatekeeper:unstable --input resources/resources.json --output resources/resources.json --json

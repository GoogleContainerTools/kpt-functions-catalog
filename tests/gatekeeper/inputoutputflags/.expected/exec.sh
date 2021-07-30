#!/usr/bin/env bash

docker run --rm -v "$(pwd)"/resources:/resources gcr.io/kpt-fn/gatekeeper:unstable --input resources/resources.yaml --output resources/resources.yaml

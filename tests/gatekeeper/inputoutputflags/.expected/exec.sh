#!/usr/bin/env bash

docker run --rm -v "$(pwd)"/resources:/resources gcr.io/kpt-fn/gatekeeper:v0.2.1 --input resources/resources.yaml --output resources/resources.yaml

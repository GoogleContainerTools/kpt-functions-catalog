#!/usr/bin/env bash

kpt fn eval --image-pull-policy never --image gcr.io/kpt-fn/render-helm-chart:v0.1 --network -- \
name=terraform \
repo=https://helm.releases.hashicorp.com \
version=1.0.0 \
releaseName=terraforming-mars \
includeCRDs=true

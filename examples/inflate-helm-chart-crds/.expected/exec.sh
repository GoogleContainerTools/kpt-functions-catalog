#!/usr/bin/env bash

kpt fn eval --image-pull-policy ifNotPresent --image gcr.io/kpt-fn/inflate-helm-chart:unstable --network -- \
name=terraform \
repo=https://helm.releases.hashicorp.com \
version=1.0.0 \
releaseName=terraforming-mars \
includeCRDs=true

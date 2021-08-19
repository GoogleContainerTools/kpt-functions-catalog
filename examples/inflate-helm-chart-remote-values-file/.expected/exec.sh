#!/usr/bin/env bash

kpt fn eval --image-pull-policy ifNotPresent --image gcr.io/kpt-fn/inflate-helm-chart:unstable \
--network \
--mount type=bind,src="$(pwd)",dst=/tmp/charts -- \
name=cert-manager \
namespace=cert-manager \
releaseName=cert-manager \
valuesFile=https://raw.githubusercontent.com/config-sync-examples/helm-components/main/cert-manager-values.yaml

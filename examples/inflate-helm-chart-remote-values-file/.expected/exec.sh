#!/usr/bin/env bash

kpt fn eval --image-pull-policy ifNotPresent --image gcr.io/kpt-fn/inflate-helm-chart:unstable \
--network \
--mount type=bind,src="$(pwd)",dst=/tmp/charts -- \
name=helloworld-chart \
releaseName=test \
valuesFile=https://raw.githubusercontent.com/GoogleContainerTools/kpt-functions-catalog/master/examples/inflate-helm-chart-local/helloworld-values/values.yaml

#!/usr/bin/env bash

kpt fn eval --image gcr.io/kpt-fn-contrib/inflate-helm-chart:unstable \
--image-pull-policy never \
--mount type=bind,src="$(pwd)"/helloworld-chart,dst=/source \
--mount type=bind,src="$(pwd)"/helloworld-values,dst=/values -- \
name=helloworld \
local-chart-path=/source \
--values=/values/values.yaml

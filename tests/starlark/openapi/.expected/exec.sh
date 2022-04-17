#! /bin/bash

kpt fn eval --image gcr.io/kpt-fn/starlark:unstable --fn-config fn-config.yaml --image-pull-policy never

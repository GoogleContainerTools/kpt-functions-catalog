#! /bin/bash

kpt fn eval --image gcr.io/kpt-fn/starlark:unstable --fn-config set-ns.yaml --image-pull-policy never

#! /bin/bash

kpt fn eval --image gcr.io/kpt-fn/starlark:unstable --image-pull-policy never -- source="$(cat set-replicas.star)" replicas=5

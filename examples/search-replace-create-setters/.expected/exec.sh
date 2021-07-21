#! /bin/bash

# shellcheck disable=SC2016
kpt fn eval --image gcr.io/kpt-fn/search-replace:v0.2 --image-pull-policy never -- \
by-path='data.**' by-value-regex='(.*)nginx.com(.*)' put-comment='kpt-set: ${1}${host}${2}'

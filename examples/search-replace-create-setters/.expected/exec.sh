#! /bin/bash

# shellcheck disable=SC2016
kpt fn eval --image gcr.io/kpt-fn/search-replace:v0.1 -- 'by-path=spec.replicas' 'put-comment=kpt-set: ${replicas}'

# shellcheck disable=SC2016
kpt fn eval --image gcr.io/kpt-fn/search-replace:v0.1 -- 'by-path=spec.**.image' 'put-comment=kpt-set: gcr.io/${image}:${tag}'

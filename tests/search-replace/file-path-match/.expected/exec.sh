#! /bin/bash

# shellcheck disable=SC2016
kpt fn eval --image gcr.io/kpt-fn/search-replace:unstable --image-pull-policy=never -- \
by-value=project-id by-file-path='**/setters.yaml' put-value=new-project

#! /bin/bash

# shellcheck disable=SC2016
kpt fn eval -i upsert-resource:unstable --image-pull-policy never --fn-config .fn-config/fn-config.yaml
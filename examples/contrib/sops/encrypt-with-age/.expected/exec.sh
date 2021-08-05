#! /bin/bash

kpt fn eval -e SOPS_IMPORT_AGE="$(cat age_keys.txt)" -i gcr.io/kpt-fn-contrib/sops:unstable --fn-config fn-config.yaml --image-pull-policy never

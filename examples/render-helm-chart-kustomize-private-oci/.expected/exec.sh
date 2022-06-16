#!/usr/bin/env bash

# kustomize 4.2.0 is preinstalled in github actions
kustomize build --enable-alpha-plugins --network > /dev/null

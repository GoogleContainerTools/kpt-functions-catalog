#!/usr/bin/env bash

kustomize build kustomization-dir --enable-alpha-plugins --network > kustomization-dir/resources.yaml

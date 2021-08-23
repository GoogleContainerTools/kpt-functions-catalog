#!/usr/bin/env bash

kustomize build --enable-alpha-plugins --network > resources.yaml

#! /bin/bash
set -eo pipefail

kpt pkg init
kpt fn render

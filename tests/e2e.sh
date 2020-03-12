#!/usr/bin/env bash
#
# E2E tests for kpt-functions-catalog.

set -eo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. >/dev/null 2>&1 && pwd)"
TAG=${TAG:-dev}
NODOCKER=${NODOCKER:-}
HELM_USAGE_SNIPPET="Render chart templates locally"

############################
# Test framework
############################

function testcase() {
  echo "testcase: ${1}"
  tmp=$(mktemp -d "/tmp/e2e.${1}.XXXXXXXX")
  cp -r "${REPO}"/example-configs "${tmp}"
  cp -r "${REPO}"/sh/test-data/* "${tmp}"
  cd "${tmp}"
}

function fail() {
  echo "FAIL: " "$@"
  exit 1
}

function assert_contains_string() {
  content="$(<"$1")"
  grep -q "$2" "$1" || fail "String $2 not contained in: ${content}"
}

function assert_dir_exists() {
  [[ -d $1 ]] || fail "Dir not exist: $1"
}

############################
# Docker Tests
############################

[[ -z ${NODOCKER} ]] || {
  echo "Skipping docker tests"
  exit 0
}

testcase "docker_helm_template_usage_too_few_args"
docker run -v "$(pwd)/helm":/source gcr.io/kpt-functions/helm-template:"${TAG}" name=too-few-args 2>err.txt || true
assert_contains_string err.txt "${HELM_USAGE_SNIPPET}"

testcase "docker_helm_template_usage_too_many_args"
docker run -v "$(pwd)/helm":/source gcr.io/kpt-functions/helm-template:"${TAG}" chart_path=/source/redis name=too-many-args extra=args_provided 2>err.txt || true
assert_contains_string err.txt "${HELM_USAGE_SNIPPET}"

testcase "docker_helm_template_correct_args"
docker run -v "$(pwd)/helm":/source gcr.io/kpt-functions/helm-template:"${TAG}" name=correct-args chart_path=/source/redis >out.yaml
assert_contains_string out.yaml "correct-args"

testcase "docker_helm_template_values_arg_sink"
docker run -v "$(pwd)/helm":/source gcr.io/kpt-functions/helm-template:"${TAG}" chart_path=/source/redis name=prod-values values_path=/source/redis/values-production.yaml |
  docker run -i -u "$(id -u)" -v "$(pwd)":/sink gcr.io/kpt-functions/write-yaml:"${TAG}" -o /dev/null -d sink_dir=/sink -d overwrite=true
assert_dir_exists redis

testcase "docker_helm_template_stdin_correct_args_sink"
docker run -v "$(pwd)/helm":/source gcr.io/kpt-functions/helm-template chart_path=/source/mongodb name=my-mongodb |
  docker run -i -v "$(pwd)/helm":/source gcr.io/kpt-functions/helm-template name=my-redis chart_path=/source/redis |
  docker run -i -u "$(id -u)" -v "$(pwd)":/sink gcr.io/kpt-functions/write-yaml:"${TAG}" -o /dev/null -d sink_dir=/sink -d overwrite=true
assert_dir_exists mongodb
assert_dir_exists redis

############################
# kpt fn Tests
############################

testcase "kpt_helm_template_imperative"
kpt fn source example-configs |
  docker run -v "$(pwd)/helm":/source gcr.io/kpt-functions/helm-template chart_path=/source/mongodb name=my-mongodb |
  docker run -i -v "$(pwd)/helm":/source gcr.io/kpt-functions/helm-template name=my-redis chart_path=/source/redis |
  kpt fn sink example-configs
assert_dir_exists ./example-configs/mongodb
assert_dir_exists ./example-configs/redis

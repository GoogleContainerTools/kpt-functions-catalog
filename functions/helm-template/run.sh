#!/usr/bin/env bash
#
# Copyright 2020 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Render chart templates locally using helm template. Display the output objects as a Kubernetes List.

set -eo pipefail

NAME="name"
CHART_PATH="chart_path"
VALUES_PATH="values_path"
USAGE=$(cat << END
Render chart templates locally using helm template.
Display the output objects as a Kubernetes List. If piped a Kubernetes List in
addition to arguments then render the chart objects into the piped list,
overwriting any chart objects that already exist in the list.

Configured using a ConfigMap with the following keys:

${NAME}: Name of helm chart.
${CHART_PATH}: Chart templates directory.
${VALUES_PATH}: [Optional] Path to values.yaml. Defaults to ${CHART_PATH}/values.yaml.

Example:

To expand a chart named 'my-chart' at '../path/to/helm/chart' using './values.yaml':

apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  annotations:
    config.k8s.io/function: |
      container:
        image:  gcr.io/kpt-functions/helm-template
    config.kubernetes.io/local-config: "true"
data:
  ${NAME}: my-chart
  ${CHART_PATH}: ../path/to/helm/chart
  ${VALUES_PATH}: ./values.yaml
END
)

err() {
  echo "[$(date +'%Y-%m-%dT%H:%M:%S%z')]: $*" >&2
}

read_arguments() {
  if [[ $# -le 1 ]] ; then
    err "${USAGE}"
    exit 1
  fi

  for arg in "$@"; do
    case $arg in
      "${NAME}="*) name="${arg#*=}" ;;
      "${CHART_PATH}="*) chartPath="${arg#*=}" ;;
      "${VALUES_PATH}="*) valuesPath="${arg#*=}" ;;
      *)
        err "${USAGE}"
        exit 1
        ;;
    esac
  done

  if [[ -z $valuesPath ]]; then
    valuesPath="${chartPath}/values.yaml"
  fi
}

create_tmp() {
  tmp=$(mktemp -d "/tmp/helm-template.XXXXXXXX")
  # If there is input through stdin then write it to tmp
  if [[ ! -t 0 ]]; then
    cat "-" | kpt fn sink "$tmp"
  fi
}

expand_helm() {
  # Overwrite files if they exist
  if ! output="$(helm template "${name}" "${chartPath}" -f "${valuesPath}" --output-dir "${tmp}" )"; then
    err "Helm template error: ${output}"
    exit 1
  fi
  kpt fn source "$tmp"
}

run_main() {
  read_arguments "$@"
  create_tmp
  expand_helm
}

run_main "$@"

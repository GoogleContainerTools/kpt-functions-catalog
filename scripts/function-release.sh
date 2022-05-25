#! /bin/bash
#
# Copyright 2021 Google LLC
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

# GIT_TAG should be set in format as
# functions/{language_name}/{function_name}/{semver}.
# e.g. "functions/go/set-namespace/v1.2.3" and "functions/ts/kubeval/v2.3.4"
# You can optionally set environment variable GCR_REGISTRY
# (e.g. "gcr.io/my-project"), then your image will be built as
# {GCR_REGISTRY}/{function_name}:{tag}.

set -euo pipefail

# If GIT_TAG is not set, just skip processing.
if [ -z "${GIT_TAG}" ];
then
  exit 0
fi

scripts_dir="$(dirname "$0")"
# git-tag-parser.sh has been shell-checked separately.
# shellcheck source=/dev/null
source "${scripts_dir}"/git-tag-parser.sh

fn_lang=$(parse_git_tag lang)
fn_name=$(parse_git_tag name)
fn_ver=$(parse_git_tag version)

if [ -d "${scripts_dir}/../functions/${fn_lang}/${fn_name}" ]; then
  cd "${scripts_dir}/../functions/${fn_lang}"
  if [ "${fn_lang}" == "go" ]; then
    make install-mdtogo
  fi
  DEFAULT_GCR="${DEFAULT_GCR:=gcr.io/kpt-fn}"
fi

if [ -d "${scripts_dir}/../contrib/functions/${fn_lang}/${fn_name}" ]; then
  cd "${scripts_dir}/../contrib/functions/${fn_lang}"
  if [ "${fn_lang}" == "go" ]; then
    make install-mdtogo
  fi
  DEFAULT_GCR="${DEFAULT_GCR:=gcr.io/kpt-fn-contrib}"
fi

case "$1" in
  build)
    CURRENT_FUNCTION="${fn_name}" TAG="${fn_ver}" DEFAULT_GCR="$DEFAULT_GCR" make func-build
    ;;
  push)
    CURRENT_FUNCTION="${fn_name}" TAG="${fn_ver}" DEFAULT_GCR="$DEFAULT_GCR" make func-push
    ;;
  *)
    echo "Usage: $0 {build|push}"
    exit 1
esac

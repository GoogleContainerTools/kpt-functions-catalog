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

set -euo pipefail

# If GIT_TAG is not set, just skip processing.
if [ -z "${GIT_TAG}" ];
then
  exit 0
fi

scripts_dir="$(dirname "$0")"
# git-tag-parser.sh has been shell-checked separately.
# shellcheck source=/dev/null
source ${scripts_dir}/git-tag-parser.sh

fn_lang=$(parse_git_tag lang)
fn_name=$(parse_git_tag name)
fn_ver=$(parse_git_tag version)

cd "${scripts_dir}"/../functions/"${fn_lang}"
CURRENT_FUNCTION="${fn_name}" TAG="${fn_ver}" make func-build

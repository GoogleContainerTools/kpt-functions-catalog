#!/bin/bash

VERNUM='0|[1-9][0-9]*'

# An example is "set-namespace/v1.2.3"
SEMVER_REGEX="\
^[a-zA-Z]*(-[a-zA-Z]+)*\/\
[vV]?($VERNUM)\\.($VERNUM)\\.($VERNUM)$"

function validate_version {
  local version=$1
  if [[ "$version" =~ $SEMVER_REGEX ]]; then
    local major=${BASH_REMATCH[2]}
    local minor=${BASH_REMATCH[3]}
    local patch=${BASH_REMATCH[4]}

    echo v$major.$minor.$patch
    echo v$major.$minor
    echo v$major
  else
    echo $version
  fi
}

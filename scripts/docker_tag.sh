#!/usr/bin/env bash
set -eo pipefail

# docker_tag.sh outputs the Docker image tag based on
# the git version and FDB version. This script is used
# by build.sh and CI workflows to ensure consistent tags.
#
# Environment variables:
#   LATEST  - If set, uses "latest" instead of git version
#   FDB_VER - FDB version (defaults to 6.2.30)

# code_version returns the latest tag for the current
# Git commit. If there are no tags associated with
# the commit then the short hash is returned.

function code_version {
  local tag=""
  if tag="$(git describe --tags 2>/dev/null)"; then
    echo "$tag"
    return 0
  fi
  git rev-parse --short HEAD
}

# fdb_version returns the version of the FDB library.
# Uses FDB_VER env var if set, otherwise defaults to 6.2.30.

function fdb_version {
  echo "${FDB_VER:-6.2.30}"
}

echo "${LATEST:-$(code_version)}_fdb.$(fdb_version)"

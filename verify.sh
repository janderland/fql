#!/usr/bin/env bash
set -exuo pipefail

# Change directory to repo root.
cd "${0%/*}"

readonly flag="${1:-}"

function docker_compose_up {
  docker compose up "$1" --exit-code-from "$1"
}

if [[ -z "$flag" ]]; then
  docker compose build verify_built
  docker_compose_up verify_built
  exit $?
elif [[ "$flag" == "--mounted" ]]; then
  docker_compose_up verify_mounted
  exit $?
else
  echo "verify.sh [--mounted]"
  exit 1
fi

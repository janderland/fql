#!/usr/bin/env bash
set -exuo pipefail

# Change directory to repo root.
cd "${0%/*}"

readonly flag="${1:-}"

if [[ -z "$flag" ]]; then
  docker compose build verify_built
  docker compose up verify_built
  exit $?
elif [[ "$flag" == "--mounted" ]]; then
  docker compose up verify_mounted
  exit $?
else
  echo "verify.sh [--mounted]"
  exit 1
fi

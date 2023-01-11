#!/usr/bin/env bash
set -eo pipefail

# Change directory to repo root.
cd "${0%/*}"

COMMAND=()
if [[ "$1" == "--generated" ]]; then
  COMMAND+=('./scripts/verify_generation.sh &&')
fi
COMMAND+=('./scripts/setup_database.sh &&')
COMMAND+=('./scripts/verify_codebase.sh')

set -x
COMMAND="${COMMAND[*]}" docker compose up verify --attach verify --exit-code-from verify

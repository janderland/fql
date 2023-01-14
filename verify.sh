#!/usr/bin/env bash
set -eo pipefail

# Change directory to repo root.
cd "${0%/*}"

while [[ $# -gt 0 ]]; do
  case $1 in
    --generated)
      GENERATED="x"
      shift
      ;;

    --build)
      BUILD="x"
      shift
      ;;
  esac
done

if [[ -n "$BUILD" ]]; then
  set -x
  COMMAND="" docker compose build
  set +x
fi

COMMAND=()
if [[ -n "$GENERATED" ]]; then
  COMMAND+=('./scripts/verify_generation.sh &&')
fi
COMMAND+=('./scripts/setup_database.sh &&')
COMMAND+=('./scripts/verify_codebase.sh')

set -x
COMMAND="${COMMAND[*]}" docker compose up verify --attach verify --exit-code-from verify

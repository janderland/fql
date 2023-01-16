#!/usr/bin/env bash
set -eo pipefail

# Change directory to repo root.
cd "${0%/*}"

# Parse the flags.
while [[ $# -gt 0 ]]; do
  case $1 in
    --generated)
      GENERATED="x"
      shift
      ;;

    *)
      echo "ERR! Invalid flag '$1'"
      exit 1
  esac
done

CMD_ARR=()
if [[ -n "$GENERATED" ]]; then
  CMD_ARR+=('./scripts/verify_generation.sh &&')
fi
CMD_ARR+=('./scripts/setup_database.sh &&')
CMD_ARR+=('./scripts/verify_codebase.sh')

set -x
COMMAND="${CMD_ARR[*]}" docker compose up verify --attach verify --exit-code-from verify

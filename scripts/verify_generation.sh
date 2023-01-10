#!/usr/bin/env bash
set -euo pipefail

# Change directory to repo root.
cd "${0%/*}/.."

# Run generation and check for differences.
go generate ./...

STATUS="$(git status --short | cut -d' ' -f3)"
if [[ -n $STATUS ]]; then
  echo "ERR! The following generated files are outdated:"
  echo "$STATUS"
  echo
  echo "Execute 'go generate ./...'"
  exit 1
fi

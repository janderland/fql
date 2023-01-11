#!/usr/bin/env bash
set -euo pipefail

# Change directory to repo root.
cd "${0%/*}/.."

STATUS="$(git status --short)"
echo "$STATUS"
if [[ -n "$STATUS" ]]; then
  echo "ERR! Generated code cannot be verified while there are uncommitted changes."
  exit 1
fi

# Run generation and check for differences.
go generate ./...

STATUS="$(git status --short)"
echo "$STATUS"
if [[ -n "$STATUS" ]]; then
  echo "ERR! The following generated files are outdated:"
  cut -d' ' -f3 <<< "$STATUS"
  echo
  echo "Execute 'go generate ./...'"
  exit 1
fi

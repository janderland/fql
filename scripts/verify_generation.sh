#!/usr/bin/env bash
set -euo pipefail

# Change directory to repo root.
cd "${0%/*}/.."

# Run generation and check for differences.
go generate ./...
if git status --short | grep .; then
  echo "ERR: Generated code outdated! Execute 'go generate ./...'"
  exit 1
fi

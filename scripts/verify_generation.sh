#!/usr/bin/env bash
set -euo pipefail

# Change directory to repo root.
cd "${0%/*}/.."

STATUS="$(git status --short)"
if [[ -n "$STATUS" ]]; then
  echo "ERR! Generated code cannot be verified while there are uncommitted changes."
  exit 1
fi

# Run generation and check for differences.
(set -x; go generate ./...)

STATUS="$(git status --short)"
if [[ -n "$STATUS" ]]; then
  echo "ERR! The following generated files are outdated:"
  cut -d' ' -f3 <<< "$STATUS"
  echo
  echo "If this is CI/CD, you'll need to execute"
  echo "'go generate ./...' on your local machine"
  echo "then commit and push the changes."
  echo
  echo "If this is your local machine then the above"
  echo "command has already been executed here."
  exit 1
fi

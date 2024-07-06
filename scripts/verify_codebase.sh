#!/usr/bin/env bash
set -euo pipefail

# Change directory to repo root.
cd "${0%/*}/.."

set -x

id
ls -ld /fdbq

# Lint shell scripts.
find . -type f -iname '*.sh' -print0 | xargs -t -0 shellcheck

# Lint Dockerfiles, if the '--no-hado' flag wasn't passed.
if [[ "${1:-}" != '--no-hado' ]]; then
  find . -type f -iname 'Dockerfile' -print0 | xargs -t -0 -n 1 hadolint
fi

# Build, lint, & test Go code.
go build ./...
golangci-lint run ./...
go test ./...

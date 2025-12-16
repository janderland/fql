#!/usr/bin/env bash
set -euo pipefail

# Change directory to repo root.
cd "${0%/*}/.."

set -x

# Lint shell scripts.
find . -type f -iname '*.sh' -not -path '*/.*' -print0 | xargs -t -0 shellcheck

# Lint Dockerfiles.
find . -type f -iname 'Dockerfile' -not -path '*/.*' -print0 | xargs -t -0 -n 1 hadolint

# Build, lint, & test Go code.
go build ./...
golangci-lint run ./...
go test ./...

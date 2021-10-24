#!/usr/bin/env bash
set -exuo pipefail

# Change directory to repo root.
cd "${0%/*}/.."

# Lint shell scripts.
find . -type f -iname '*.sh' -print0 | xargs -0 shellcheck

# Lint Dockerfiles.
find . -type f -iname 'Dockerfile' -print0 | xargs -0 -n 1 hadolint

# build, lint, & test Go code.
go build ./...
golangci-lint run ./...
go test ./...

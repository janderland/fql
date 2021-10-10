#!/usr/bin/env bash
set -exuo pipefail

cd "${0%/*}/.."

# Lint shell scripts.
find . -type f -iname '*.sh' -print0 | xargs -0 shellcheck

# Lint Dockerfiles.
find . -type f -iname 'Dockerfile' -print0 | xargs -0 -n 1 hadolint

# Lint, build, & test Go code.
golangci-lint run
go build ./...
go test ./...

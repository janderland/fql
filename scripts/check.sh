#!/usr/bin/env bash
set -exuo pipefail

cd "${0%/*}/.."
find . -type f -iname '*.sh' -print0 | xargs -0 shellcheck
find . -type f -iname 'Dockerfile' -print0 | xargs -0 -n 1 hadolint
golangci-lint run
go build ./...
go test ./...


#!/usr/bin/env bash
set -exuo pipefail

cd "${0%/*}"
find . -type f -iname '*.sh' -print0 | xargs -0 shellcheck
golangci-lint run
go build ./...
go test ./...


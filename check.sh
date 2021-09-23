#!/usr/bin/env bash
set -exuo pipefail

cd "${0%/*}"
golangci-lint run
go build ./...
go test ./...

#!/usr/bin/env bash
set -exuo pipefail

# Change directory to repo root.
cd "${0%/*}"

docker compose up verify --attach verify --exit-code-from verify

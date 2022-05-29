#!/usr/bin/env bash
set -exuo pipefail

# Change directory to repo root.
cd "${0%/*}"

docker compose up verify --exit-code-from verify

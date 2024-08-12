#!/usr/bin/env bash
set -euo pipefail

# Change directory to repo root.
cd "${0%/*}/.."

pandoc \
  --no-highlight --toc \
  --template ./web/index.tmpl \
  --output ./web/index.html \
  ./web/index.md

#!/usr/bin/env bash
set -euo pipefail

# Change directory to repo root.
cd "${0%/*}/.."

pandoc \
  --no-highlight --toc \
  --template ./docs/index.tmpl \
  --output ./docs/index.html \
  ./docs/index.md

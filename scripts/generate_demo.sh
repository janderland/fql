#!/usr/bin/env bash
set -euo pipefail

# Change directory to repo root.
cd "${0%/*}/.."

vhs ./vhs/demo.tape --output ./vhs/demo.gif

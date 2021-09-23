#!/usr/bin/env bash
set -exuo pipefail

cd "${0%/*}"

FDBQ_FDB_NAME="fdbq_fdb"
export FDBQ_FDB_NAME

FDBQ_SRC_DIR="$(pwd)"
export FDBQ_SRC_DIR

docker compose up \
  --abort-on-container-exit \
  --always-recreate-deps \
  --renew-anon-volumes \
  --force-recreate \
  check


#!/usr/bin/env bash
set -euo pipefail

# Expand the contents of the first argument and store
# for later use as the contents of the cluster file.
# Shift the rest of the arguments up if there ary any.
CLUSTER_FILE="$(eval echo "${1:-}")"
echo "shim.sh - cluster file: $CLUSTER_FILE"
shift || true

# Create the cluster file and run FDBQ with the
# remaining arguments.
echo "$CLUSTER_FILE" > /etc/foundationdb/fdb.cluster
/fdbq "$@"

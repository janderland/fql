#!/usr/bin/env bash
set -exuo pipefail

# Obtain the IP for FDB from the given hostname.
FDB_IP=$(getent hosts "$1" | awk '{print $1}')

# Create the FDB cluster file.
echo "docker:docker@${FDB_IP}:4500" > /etc/foundationdb/fdb.cluster

# If this is a fresh instance of FDB, configure the database.
function not_configured {
  fdbcli --exec 'status json' | jq -e '.cluster.messages[] | select(.name | contains("unreadable_configuration"))'
}
if not_configured; then
  fdbcli --exec "configure new single memory"
fi


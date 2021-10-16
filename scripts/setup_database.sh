#!/usr/bin/env bash
set -exuo pipefail

# The first argument is the hostname of the FDB container.
FDB_HOSTNAME=${1:-fdb}

# The second argument is the description & ID of the FDB cluster.
# https://apple.github.io/foundationdb/administration.html#cluster-file-format
FDB_DESCRIPTION_ID=${2:-docker:docker}

# Obtain the IP for FDB from the given hostname.
FDB_IP=$(getent hosts "$FDB_HOSTNAME" | awk '{print $1}')

# Create the FDB cluster file.
echo "${FDB_DESCRIPTION_ID}@${FDB_IP}:4500" > /etc/foundationdb/fdb.cluster

# This function returns error code '0' if FDB isn't configured.
function not_configured {
  fdbcli --exec 'status json' | jq -e '.cluster.messages[] | select(.name | contains("unreadable_configuration"))'
}

# If this is a new instance of FDB, configure the database.
# https://apple.github.io/foundationdb/administration.html#re-creating-a-database
if not_configured; then
  fdbcli --exec "configure new single memory"
fi

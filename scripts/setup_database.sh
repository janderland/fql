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

# Search for the "unreadable_configuration" message in the cluster's status. This message
# would let us know that the database hasn't been initialized.
JQ_CODE=$(
  jq -e '.cluster.messages[] | select(.name | contains("unreadable_configuration"))' \
    <(fdbcli --exec 'status json') >&2
  echo $?
)

# jq should only return codes between 0 & 4 inclusive. Our particular query never
# returns 'null' or 'false', so we shouldn't see code 1. Codes 2 & 3 occur on
# system & compile errors respectively, so the only valid codes are 0 & 4.
# https://stedolan.github.io/jq/manual/#Invokingjq
if [[ $JQ_CODE -lt 0 || ( $JQ_CODE -gt 0 && $JQ_CODE -lt 4 ) || $JQ_CODE -gt 4 ]]; then
  exit "$JQ_CODE"
fi

# If this is a new instance of FDB, configure the database.
# https://apple.github.io/foundationdb/administration.html#re-creating-a-database
if [[ $JQ_CODE -eq 0 ]]; then
  fdbcli --exec "configure new single memory"
fi

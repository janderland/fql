#!/usr/bin/env bash
set -u

# The first argument is the hostname of the FDB container.
FDB_HOSTNAME=${1:-fdb}
echo "FDB_HOSTNAME=$FDB_HOSTNAME"

# The second argument is the description & ID of the FDB cluster.
# https://apple.github.io/foundationdb/administration.html#cluster-file-format
FDB_DESCRIPTION_ID=${2:-docker:docker}
echo "FDB_DESCRIPTION_ID=$FDB_DESCRIPTION_ID"

# Obtain the IP for FDB from the given hostname.
FDB_IP=$(getent hosts "$FDB_HOSTNAME" | awk '{print $1}')
echo "FDB_IP=$FDB_IP"

# This variable is recognised by fdbcli and ensures that a fdb.cluster file in
# the current directory won't interfere with this script.
export FDB_CLUSTER_FILE="/etc/foundationdb/fdb.cluster"

# Create the FDB cluster file.
echo "${FDB_DESCRIPTION_ID}@${FDB_IP}:4500" > $FDB_CLUSTER_FILE
echo "FDB_CLUSTER_FILE: $(cat $FDB_CLUSTER_FILE)"

# Search for the "unreadable_configuration" message in the cluster's status. This message
# would let us know that the database hasn't been initialized.
jq -e '.cluster.messages[] | select(.name | contains("unreadable_configuration"))' <(fdbcli --exec 'status json') >&2
JQ_CODE=$?

# jq should only return codes between 0 & 4 inclusive. Our particular query never
# returns 'null' or 'false', so we shouldn't see code 1. Codes 2 & 3 occur on
# system & compile errors respectively, so the only valid codes are 0 & 4. If the
# code is not 0 or 4 then something unexpected happened so return early.
# https://stedolan.github.io/jq/manual/#Invokingjq
if [[ $JQ_CODE -lt 0 || ( $JQ_CODE -gt 0 && $JQ_CODE -lt 4 ) || $JQ_CODE -gt 4 ]]; then
  echo "ERR! jq exited with code $JQ_CODE."
  exit "$JQ_CODE"
fi

# If this is a new instance of FDB, configure the database.
# https://apple.github.io/foundationdb/administration.html#re-creating-a-database
if [[ $JQ_CODE -eq 0 ]]; then
  set -x
  fdbcli --exec "configure new single memory"
  set +x
fi

# If we make it this far then jq should have returned code 4
# which means there's no need to configure the database and
# the script can exit.

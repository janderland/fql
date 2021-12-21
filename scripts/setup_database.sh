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

# Query for the status of the FDB cluster.
status=$(fdbcli --exec 'status json')

# Search for the "unreadable_configuration" message in the cluster's status. This message would let us
# know that the database hasn't been initialized. With the '-e' flag, jq will return 0 if the message
# is found and 1 if it isn't. Any other error code should be treated as a script failure.
# https://stedolan.github.io/jq/manual/#Invokingjq
# NOTE: This command is run in a sub-shell so 'set -e' doesn't cause an immediate exit.
JQ_CODE=$(
  jq -e '.cluster.messages[] | select(.name | contains("unreadable_configuration"))' <(echo "$status") >&2
  echo $?
)
if [[ $JQ_CODE -gt 1 ]] || [[ $JQ_CODE -lt 0 ]]; then
  exit $JQ_CODE
fi

# If this is a new instance of FDB, configure the database.
# https://apple.github.io/foundationdb/administration.html#re-creating-a-database
if $JQ_CODE -eq 0 then
  fdbcli --exec "configure new single memory"
  exit $?
fi

#!/usr/bin/env bash
set -exuo pipefail

FDB_IP=$(getent hosts "$1" | awk '{print $1}')
echo "docker:docker@${FDB_IP}:4500" > /etc/foundationdb/fdb.cluster
fdbcli --exec "configure new single memory"


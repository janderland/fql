#!/usr/bin/env bash
set -exuo pipefail

DIR="${0%/*}"
TAG="janderland/fdbq:$(git rev-parse --short HEAD)"
ROOT="$(cd "$DIR"/../..; pwd)"
WORKDIR="/fdbq"

docker run -v "$ROOT":$WORKDIR -w $WORKDIR janderland/fdbq-build:latest go build
docker build -t "$TAG" -f "$DIR"/Dockerfile "$ROOT"
docker push "$TAG"

#!/usr/bin/env bash
set -exuo pipefail

DIR="${0%/*}"
TAG="janderland/fdbq:$(git rev-parse --short HEAD)"
ROOT="$(cd "$DIR"/../..; pwd)"

docker run -v "$ROOT":"$ROOT" -w "$ROOT" janderland/fdbq-build:latest go build "$ROOT"
docker build -t "$TAG" -f "$DIR"/Dockerfile "$ROOT"
docker push "$TAG"

#!/usr/bin/env bash
set -exuo pipefail

DIR="${0%/*}"
TAG="janderland/fdbq-build:latest"

docker build -t "$TAG" - < "$DIR"/Dockerfile
# docker push "$TAG"

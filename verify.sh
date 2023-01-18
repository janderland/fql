#!/usr/bin/env bash
set -eo pipefail


# Change directory to repo root.

cd "${0%/*}"


# Parse the flags.

while [[ $# -gt 0 ]]; do
  case $1 in
    --generated)
      VERIFY_GENERATION="x"
      shift 1
      ;;

    --verify)
      VERIFY_CODEBASE="x"
      shift 1
      ;;

    --build)
      for service in $(echo "$2" | tr "," "\n"); do
        case $service in
          build)
            BUILD_BUILD_CONTAINER="x"
            ;;
          fdbq)
            BUILD_FDBQ_CONTAINER="x"
            ;;
          *)
            echo "ERR! Invalid build target '$service'"
            exit 1
            ;;
        esac
      done
      shift 2
      ;;

    --)
      shift 1
      FDBQ_ARGS=("$@")
      shift $#
      ;;

    *)
      echo "ERR! Invalid flag '$1'"
      exit 1
  esac
done


# Build variables required by the docker compose command.

BUILD_TASKS=()

if [[ -n "$VERIFY_GENERATION" ]]; then
  BUILD_TASKS+=('./scripts/verify_generation.sh &&')
fi

if [[ -n "$VERIFY_CODEBASE" ]]; then
  BUILD_TASKS+=('./scripts/setup_database.sh &&')
  BUILD_TASKS+=('./scripts/verify_codebase.sh')
fi

export BUILD_COMMAND="${BUILD_TASKS[*]}"
echo "BUILD_COMMAND=${BUILD_TASKS[*]}"

export FDBQ_COMMAND="${FDBQ_ARGS[*]}"
echo "FDBQ_COMMAND=${FDBQ_ARGS[*]}"

export FDBQ_TAG="latest"
echo "FDBQ_TAG=latest"


# Run the requested commands.

if [[ -n "$BUILD_BUILD_CONTAINER" ]]; then
  (set -x;
    docker compose build build
  )
fi

if [[ -n "$BUILD_COMMAND" ]]; then
  (set -x;
    docker compose up build --attach build --exit-code-from build
  )
fi

if [[ -n "$BUILD_FDBQ_CONTAINER" ]]; then
  (set -x;
    docker compose build fdbq
  )
fi

if [[ -n "$FDBQ_COMMAND" ]]; then
  (set -x;
    docker compose up fdbq --attach fdbq --exit-code-from fdbq
  )
fi

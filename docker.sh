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


# Define helper functions.

# join_by joins the elements of the $2 array into a
# single string, placing $1 between each element.

function join_by {
  local sep="$1"
  local out="$2"
  if shift 2; then
    for arg in "$@"; do
      out="$out $sep $arg"
    done
  fi
  echo "$out"
}


# escape_quotes adds an extra layer of single quotes
# around it's arguments. Any single quotes included
# in the arguments are escaped with backslashes.
#
# TODO: Figure out a way around this.
# We use this function on the fdbq args passed during
# the ./docker -- <args> usecase. While passing these
# args as an environment variable into the Docker
# compose file, they seem to be evaluated by a shell
# and stripped of their first layer of quotes. This
# function protects against that.

function escape_quotes {
  out=()
  for arg in "$@"; do
    out+=("$(printf "'%s'" "${arg//'/\\'}")")
  done
  echo "${out[@]}" 
}


# Build variables required by the docker compose command.

BUILD_TASKS=()

if [[ -n "$VERIFY_GENERATION" ]]; then
  BUILD_TASKS+=('./scripts/verify_generation.sh')
fi

if [[ -n "$VERIFY_CODEBASE" ]]; then
  BUILD_TASKS+=('./scripts/setup_database.sh')
  BUILD_TASKS+=('./scripts/verify_codebase.sh')
fi

BUILD_COMMAND="$(join_by ' && ' "${BUILD_TASKS[@]}")"
echo "BUILD_COMMAND=${BUILD_COMMAND}"
export BUILD_COMMAND

FDBQ_COMMAND="$(escape_quotes "${FDBQ_ARGS[@]}")"
echo "FDBQ_COMMAND=${FDBQ_COMMAND}"
export FDBQ_COMMAND

FDBQ_TAG="latest"
echo "FDBQ_TAG=latest"
export FDBQ_TAG


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

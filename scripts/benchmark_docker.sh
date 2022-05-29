#!/usr/bin/env bash
set -euo pipefail

echo "This script was used to compared CI Docker setups."
echo "One setup mounted the root dir on to the container."
echo "The other setup built the root dir into a new image"
echo "every time CI was run. Because the MacOS integration"
echo "for mounting Docker volumes was so bad, the latter"
echo "was once the faster setup. The MacOS integration"
echo "has improved so there is no need for the built-in"
echo "setup nor this benchmarking script."
exit 1

# Change directory to repo root.
cd "${0%/*}/.."

SCRIPT_NAME="$0"
OUT_FILE="./benchmark_out.txt"
BUILT_IN_FILE=""
MOUNTED_FILE=""
WARM_UP_COUNT=1
BENCHMARK_COUNT=1

# help echos a CLI flags description to stdout.
function print_help {
  echo "$SCRIPT_NAME [--built-in CACHE_FILE] [--mounted CACHE_FILE] [--warm-ups WARM_UP_COUNT] [--iterations ITERATION_COUNT] [--out OUTPUT_FILE]"
}

# log echos a prefix followed by the first
# argument to stderr.
function log {
  echo "$SCRIPT_NAME | $1" >&2
}

# perform_benchmark stops all docker compose containers,
# runs warm up iterations, and then collects benchmarks
# using the time command. The 1st argument is a file
# where the results from the time command are stored.
# The rest of the arguments is the command to benchmark. 
function perform_benchmark {
  local out=$1
  shift

  # Reset docker environment.
  docker compose down -v

  # Warm up the system with some pre-runs.
  for i in $(seq $WARM_UP_COUNT); do
    log "Pre-Run $i"
    "$@"
  done

  # Run the benchmark iterations. During each iteration,
  # run the command using the time utility. The final 3
  # lines of console output, which contain the timing
  # results for the iteration, are saved into the
  # output file.
  for i in $(seq $BENCHMARK_COUNT); do
    file=$(mktemp)
    log "Iteration $i: $file"
    { time "$@"; } 2>&1 | tee "$file"
    tail -n 3 "$file" >> "$out"
    echo "" >> "$out"
  done
}

# collect_real_durations reads a file produced by
# perform_benchmark as it's only argument and echos
# all the real time values.
function collect_real_durations {
  awk -F '\t' -f <(cat << prog
# The perform_benchmark output contains
# 3 lines per iteration. Only use the
# lines for the "real" duration rather
# than "user" or "sys".
/real/ {
  # This program requires Awk's field
  # separator to be the tab character.
  # If that's the case, the 2nd field
  # will be the duration string.
  dur = \$2

  # Split the duration string into
  # separate minutes and seconds.
  split(dur, parts, /m/);

  # The last command got rid of the 'm'
  # character. This one gets rid of the
  # 's'. Now both expressions can be
  # treated as numbers.
  parts[2] = substr(parts[2], 0, length(parts[2])-1);

  # Print the total duration in seconds.
  print (parts[1] * 60) + parts[2];
}
prog
  ) "$1"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --out)        OUT_FILE="$2";        shift 2 ;;
    --built-in)   BUILT_IN_FILE="$2";   shift 2 ;;
    --mounted)    MOUNTED_FILE="$2";    shift 2 ;;
    --warm-ups)   WARM_UP_COUNT="$2";   shift 2 ;;
    --iterations) BENCHMARK_COUNT="$2"; shift 2 ;;
    *)            print_help;           exit  1 ;;
  esac
done

if [ -e "$OUT_FILE" ]; then
  log "output file already exists: $OUT_FILE"
  exit 1
fi

BUILT_IN_FILE="${BUILT_IN_FILE:-$(mktemp)}"
log "Built-In Benchmarks: $BUILT_IN_FILE"
if [ -s "$BUILT_IN_FILE" ]; then
  log "Using cache..."
else
  perform_benchmark "$BUILT_IN_FILE" './verify.sh' 
fi

MOUNTED_FILE="${MOUNTED_FILE:-$(mktemp)}"
log "Mounted Benchmarks: $MOUNTED_FILE"
if [ -s "$MOUNTED_FILE" ]; then
  log "Using cache..."
else
  perform_benchmark "$MOUNTED_FILE" './verify.sh' '--mounted' 
fi

# Below, the SC2207 check can be ignored because collect_real_durations
# echos a list of items separated by whitespace, which works with bash's
# word splitting.

# shellcheck disable=SC2207
built_in_values=( $(collect_real_durations "$BUILT_IN_FILE") )

# shellcheck disable=SC2207
mounted_values=( $(collect_real_durations "$MOUNTED_FILE") )

if [ ${#built_in_values[@]} -ne ${#mounted_values[@]} ]; then
  log "number of data points isn't equal: ${#built_in_values[@]} != ${#mounted_values[@]}"
  exit 1
fi

echo "Built-In,Mounted" >> "$OUT_FILE"
for ((i = 0; i < ${#built_in_values[@]}; i++)); do
  echo "${built_in_values[$i]},${mounted_values[$i]}" >> "$OUT_FILE"
done

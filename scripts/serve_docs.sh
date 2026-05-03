#!/usr/bin/env bash
set -euo pipefail

# Change directory to repo root.
cd "${0%/*}/.."

CHOKIDAR_PID=/tmp/fql-serve-docs-chokidar.pid
BROWSERSYNC_PID=/tmp/fql-serve-docs-browsersync.pid
CHOKIDAR_LOG=/tmp/fql-serve-docs-chokidar.log
BROWSERSYNC_LOG=/tmp/fql-serve-docs-browsersync.log

# Substrings we expect to find in the recorded PID's command line. Matching
# before acting guards against PID recycling after an unclean exit.
CHOKIDAR_MATCH='chokidar'
BROWSERSYNC_MATCH='browser-sync'

# Returns 0 iff the PID in $1 is alive AND its command line contains $2.
pid_is_ours() {
  local pidfile=$1 match=$2 pid
  [[ -f "$pidfile" ]] || return 1
  pid=$(cat "$pidfile" 2>/dev/null) || return 1
  [[ -n "$pid" ]] || return 1
  kill -0 "$pid" 2>/dev/null || return 1
  ps -ww -o command= -p "$pid" 2>/dev/null | grep -q -- "$match"
}

# Print $1 and every descendant PID, depth-first.
pid_tree() {
  local pid=$1 child
  echo "$pid"
  for child in $(pgrep -P "$pid" 2>/dev/null); do
    pid_tree "$child"
  done
}

# Send signal $2 (default TERM) to $1 and all descendants.
kill_tree() {
  local root=$1 sig=${2:-TERM} pid
  while IFS= read -r pid; do
    kill -"$sig" "$pid" 2>/dev/null || true
  done < <(pid_tree "$root")
}

# Terminate the process tree rooted at the PID in $1. Escalates TERM -> KILL
# after a grace period. Only removes the pidfile when the root is actually
# gone, so a wedged process stays visible to the next --stop invocation.
terminate() {
  local pidfile=$1 pid i
  pid=$(cat "$pidfile")
  kill_tree "$pid" TERM
  for i in 1 2 3 4 5; do
    kill -0 "$pid" 2>/dev/null || break
    sleep 1
  done
  if kill -0 "$pid" 2>/dev/null; then
    kill_tree "$pid" KILL
    sleep 1
  fi
  if kill -0 "$pid" 2>/dev/null; then
    echo "WARN: PID $pid survived SIGKILL; leaving $pidfile in place." >&2
    return 1
  fi
  rm -f "$pidfile"
}

stop() {
  local any=0 pair pidfile match
  for pair in \
    "$CHOKIDAR_PID:$CHOKIDAR_MATCH" \
    "$BROWSERSYNC_PID:$BROWSERSYNC_MATCH"
  do
    pidfile=${pair%:*}
    match=${pair#*:}
    if pid_is_ours "$pidfile" "$match"; then
      any=1
      terminate "$pidfile" || true
    else
      # Stale or unrelated PID — safe to discard.
      rm -f "$pidfile"
    fi
  done
  if [[ $any -eq 1 ]]; then
    echo "Stopped."
  else
    echo "Not running."
  fi
}

show_log() {
  local file
  case "${1:-}" in
    chokidar)    file=$CHOKIDAR_LOG ;;
    browsersync) file=$BROWSERSYNC_LOG ;;
    *) echo "Usage: $0 --log <chokidar|browsersync>" >&2; exit 1 ;;
  esac
  if [[ ! -f "$file" ]]; then
    echo "Log file not found: $file" >&2
    exit 1
  fi
  "${PAGER:-less}" "$file"
}

case "${1:-}" in
  --stop) stop; exit 0 ;;
  --log)  shift; show_log "${1:-}"; exit 0 ;;
  "")     ;;
  *)      echo "Usage: $0 [--stop | --log <chokidar|browsersync>]" >&2; exit 1 ;;
esac

if pid_is_ours "$CHOKIDAR_PID" "$CHOKIDAR_MATCH" \
   || pid_is_ours "$BROWSERSYNC_PID" "$BROWSERSYNC_MATCH"; then
  echo "Already running. Use '$0 --stop' first." >&2
  exit 1
fi

# Drop any stale pidfiles (dead PID, or PID reused by an unrelated process).
rm -f "$CHOKIDAR_PID" "$BROWSERSYNC_PID"

npx chokidar-cli "docs/index.md" "docs/index.tmpl" -c "./build.sh --docs" >"$CHOKIDAR_LOG" 2>&1 &
echo $! > "$CHOKIDAR_PID"

npx browser-sync start --server docs --files "docs/index.html" >"$BROWSERSYNC_LOG" 2>&1 &
echo $! > "$BROWSERSYNC_PID"

echo "Started."
echo "  chokidar     PID $(cat "$CHOKIDAR_PID")  log: $CHOKIDAR_LOG"
echo "  browser-sync PID $(cat "$BROWSERSYNC_PID")  log: $BROWSERSYNC_LOG"
echo "Stop with: $0 --stop"

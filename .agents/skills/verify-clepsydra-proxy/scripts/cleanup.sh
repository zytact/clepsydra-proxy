#!/usr/bin/env bash
# Stops the instance launch.sh started and removes its run dir. Evidence in
# ~/.cache/verify-clepsydra-proxy/evidence stays.
set -uo pipefail
STATE_DIR="${VERIFY_STATE_DIR:-/tmp/verify-clepsydra-proxy}"
if [ -f "$STATE_DIR/proxy.pid" ]; then
  pid="$(cat "$STATE_DIR/proxy.pid")"
  kill -- -"$pid" 2>/dev/null || kill "$pid" 2>/dev/null
  for _ in $(seq 1 40); do
    kill -0 "$pid" 2>/dev/null || break
    sleep 0.25
  done
  kill -0 "$pid" 2>/dev/null && kill -9 "$pid" 2>/dev/null
  rm -f "$STATE_DIR/proxy.pid"
fi
rm -rf "$STATE_DIR/run" "$STATE_DIR/session.env" "$STATE_DIR/proxy.out"
echo "cleaned up"

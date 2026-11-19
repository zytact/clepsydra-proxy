#!/usr/bin/env bash
set -euo pipefail
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO="$(cd "$HERE/../../../.." && pwd)"
STATE_DIR="${VERIFY_STATE_DIR:-/tmp/verify-clepsydra-proxy}"
PORT="${VERIFY_PORT:-18317}"

if [ -f "$STATE_DIR/session.env" ]; then
  echo "A session already exists in $STATE_DIR. Run cleanup.sh first." >&2
  exit 1
fi
mkdir -p "$STATE_DIR/bin" "$STATE_DIR/run/auths" "$STATE_DIR/home"

(cd "$REPO" && go build -o "$STATE_DIR/bin/cli-proxy-api" ./cmd/server)

KEY="verify-$(head -c 12 /dev/urandom | od -An -tx1 | tr -d ' \n')"
cat >"$STATE_DIR/run/config.yaml" <<YAML
config-version: 8
server:
  host: "127.0.0.1"
  port: $PORT
management:
  allow-remote: false
  secret-key: "$KEY"
  disable-control-panel: true
access:
  api-keys:
    - "verify-client-key"
oauth:
  auth-dir: "$STATE_DIR/run/auths"
observability:
  logs:
    logging-to-file: true
YAML

(cd "$STATE_DIR/run" && HOME="$STATE_DIR/home" setsid bash -c 'echo $$ >"$0/proxy.pid"; exec "$1" --config "$2"' \
  "$STATE_DIR" "$STATE_DIR/bin/cli-proxy-api" "$STATE_DIR/run/config.yaml" \
  >"$STATE_DIR/proxy.out" 2>&1 &)

for _ in $(seq 1 60); do
  code="$(curl -s -o /dev/null -w '%{http_code}' -H "Authorization: Bearer $KEY" "http://127.0.0.1:$PORT/v8/management/config" || true)"
  [ "$code" = "200" ] && break
  sleep 0.5
done
[ "${code:-}" = "200" ] || {
  kill "$(cat "$STATE_DIR/proxy.pid")" 2>/dev/null || true
  rm -f "$STATE_DIR/proxy.pid"
  echo "Proxy did not answer and was stopped, see $STATE_DIR/proxy.out" >&2
  exit 1
}

{
  echo "BASE='http://127.0.0.1:$PORT'"
  echo "KEY='$KEY'"
  echo "RUN_DIR='$STATE_DIR/run'"
  echo "AUTH_DIR='$STATE_DIR/run/auths'"
  echo "HOME_DIR='$STATE_DIR/home'"
} >"$STATE_DIR/session.env"
echo "ready: http://127.0.0.1:$PORT  logs $STATE_DIR/proxy.out"

#!/usr/bin/env bash
set -euo pipefail
STATE_DIR="${VERIFY_STATE_DIR:-/tmp/verify-clepsydra-proxy}"
source "$STATE_DIR/session.env"
name="$1"; shift
expires="$(date -u -d '+8 hours' +%Y-%m-%dT%H:%M:%SZ)"
base_url=""
while [ $# -gt 0 ]; do
  case "$1" in
    --expired) expires="$(date -u -d '-1 hour' +%Y-%m-%dT%H:%M:%SZ)" ;;
    --base-url) base_url="\"base_url\": \"$2\","; shift ;;
    *) echo "unknown flag $1" >&2; exit 2 ;;
  esac
  shift
done
cat >"$AUTH_DIR/antigravity-$name.json" <<JSON
{
  "type": "antigravity",
  "email": "$name@example.invalid",
  $base_url
  "project_id": "fake-project-$name",
  "access_token": "fake-access-$name",
  "refresh_token": "fake-refresh-$name",
  "expired": "$expires",
  "timestamp": $(date +%s000)
}
JSON
echo "$AUTH_DIR/antigravity-$name.json"

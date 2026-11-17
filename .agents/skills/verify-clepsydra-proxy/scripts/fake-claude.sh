#!/usr/bin/env bash
# Writes a fake Claude account into the isolated auth dir. Its tokens are
# placeholders, so any upstream call is rejected; that is the point.
#   fake-claude.sh <name> [--expired] [--offline]
#   --expired  the access token expired an hour ago, so the proxy refreshes it
#   --offline  routes the account through a closed local port, so every
#              upstream dial fails the way it does before DNS is up
set -euo pipefail
STATE_DIR="${VERIFY_STATE_DIR:-/tmp/verify-clepsydra-proxy}"
source "$STATE_DIR/session.env"
name="$1"; shift
expires="$(date -u -d '+8 hours' +%Y-%m-%dT%H:%M:%SZ)"
proxy=""
for flag in "$@"; do
  case "$flag" in
    --expired) expires="$(date -u -d '-1 hour' +%Y-%m-%dT%H:%M:%SZ)" ;;
    --offline) proxy='"proxy_url": "http://127.0.0.1:9",' ;;
    *) echo "unknown flag $flag" >&2; exit 2 ;;
  esac
done
cat >"$AUTH_DIR/claude-$name.json" <<JSON
{
  "type": "claude",
  "email": "$name@example.invalid",
  $proxy
  "access_token": "fake-access-$name",
  "refresh_token": "fake-refresh-$name",
  "expired": "$expires",
  "last_refresh": "$(date -u -d '-2 hours' +%Y-%m-%dT%H:%M:%SZ)"
}
JSON
echo "$AUTH_DIR/claude-$name.json"

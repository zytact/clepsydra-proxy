#!/usr/bin/env bash
set -euo pipefail
STATE_DIR="${VERIFY_STATE_DIR:-/tmp/verify-clepsydra-proxy}"
source "$STATE_DIR/session.env"
file="$HOME_DIR/.gemini/antigravity-cli/antigravity-oauth-token"
mkdir -p "$(dirname "$file")"
case "${1:-}" in
  "")
    expiry="$(date -u -d '+1 hour' +%Y-%m-%dT%H:%M:%S.000000Z)"
    cat >"$file" <<JSON
{"token":{"access_token":"fake-cli-access","token_type":"Bearer","refresh_token":"fake-cli-refresh","expiry":"$expiry","expires_in":3600},"auth_method":"consumer","id_token_email":"cli@example.invalid"}
JSON
    ;;
  --corrupt) printf '{"token":{"access_token":"fake-cli-access"}}' >"$file" ;;
  --remove) rm -f "$file" ;;
  *) echo "unknown flag $1" >&2; exit 2 ;;
esac
echo "$file"

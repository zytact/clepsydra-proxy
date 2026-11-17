#!/usr/bin/env bash
set -euo pipefail
STATE_DIR="${VERIFY_STATE_DIR:-/tmp/verify-clepsydra-proxy}"
source "$STATE_DIR/session.env"
method="$1"; path="$2"; shift 2
curl -sS -X "$method" -H "Authorization: Bearer $KEY" "$BASE/v8/management$path" "$@"
echo

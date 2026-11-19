# Quota history

The background recorder that reads every enabled Claude, Codex and Antigravity account's usage windows on a timer and keeps a JSON line per account per poll on disk.

## Sub-features

- The recorder starts 30 seconds after the proxy and polls every 5 minutes.
- Each poll appends one JSON line per enabled Claude, Codex or Antigravity account to `quota-history/<UTC date>.jsonl`, in the folder that holds the config file.
- A failed poll stores the reason in `error` and no windows, so a rejected or expired token still leaves a reading.
- Disabled accounts are skipped, and so is every other provider.
- Antigravity readings come from a POST to `/v1internal:retrieveUserQuotaSummary` with `{"project": <project_id>}` (or `{}` without one). There is one window per enabled bucket: `id` is the `bucketId`, `window_seconds` is 18000 for `5h` and 604800 for `weekly`, and `scope` is `model`. Disabled buckets are dropped.
- Antigravity tries the daily host, then prod, unless the account sets `base_url`, which is then the only host. A 401 or 403 stops the fallback, and the first host's error is the one stored.
- Day files older than 21 days are pruned.
- `GET /v8/management/quota/history?since=<RFC 3339 time>` returns `interval_seconds`, `since` and `readings` oldest first, defaulting to the last 15 days.

## How to get to it (user POV)

- The panel's Pool Quota page reads `GET /v8/management/quota/history` and draws its graphs from the readings.
- The JSONL files are the durable copy, so the graphs survive a restart.

## Driving it with api.sh

```bash
$P/fake-claude.sh alpha
$P/fake-claude.sh beta --expired
# the first poll is 30 seconds after launch
$P/api.sh GET /quota/history
ls /tmp/verify-clepsydra-proxy/run/quota-history/
$P/api.sh GET '/quota/history?since=yesterday'
```

Proof: the response has `interval_seconds: 300`; alpha has `error: "usage request returned status 401"` and beta has `error: "access token expired"`, each with `at`, `auth_id` and `auth_index`; the JSONL file under `run/quota-history/` holds the same lines; `since=yesterday` answers 400 `since must be an RFC 3339 time`.

Disabling an account needs the next poll to show the skip:

```bash
python3 - <<'PY'
import json; p='/tmp/verify-clepsydra-proxy/run/auths/claude-alpha.json'
d=json.load(open(p)); d['disabled']=True; json.dump(d, open(p,'w'))
PY
# wait for the next 5-minute poll, then
$P/api.sh GET /quota/history
```

Proof: no reading for `claude-alpha.json` after the poll that follows the edit. `/credentials` shows the file as `disabled` immediately, but the recorder only acts on its own schedule.

## Driving Antigravity with a local stub

`quota-stub.py <port> [status] [request-log]` answers `retrieveUserQuotaSummary` with four enabled buckets and one disabled one, or with the given error status, and logs each request's path, body and user agent. `fake-antigravity.sh <name> [--expired] [--base-url URL]` writes a fake Antigravity account with `project_id: fake-project-<name>`. Write the accounts within 30 seconds of launch to catch the first poll.

```bash
STATE_DIR=${VERIFY_STATE_DIR:-/tmp/verify-clepsydra-proxy}
mkdir -p "$STATE_DIR/stub"
(setsid python3 -I $P/quota-stub.py 18491 200 "$STATE_DIR/stub/ok.log" >/dev/null 2>&1 & echo $! >"$STATE_DIR/stub/ok.pid")
(setsid python3 -I $P/quota-stub.py 18492 401 "$STATE_DIR/stub/denied.log" >/dev/null 2>&1 & echo $! >"$STATE_DIR/stub/denied.pid")
$P/fake-antigravity.sh ok --base-url http://127.0.0.1:18491
$P/fake-antigravity.sh denied --base-url http://127.0.0.1:18492
$P/fake-antigravity.sh public
# after the next poll
$P/api.sh GET /quota/history
cat "$STATE_DIR/stub/"*.log
kill "$(cat "$STATE_DIR/stub/ok.pid")" "$(cat "$STATE_DIR/stub/denied.pid")"   # before cleanup.sh
```

Proof:
- `antigravity-ok.json` has windows `gemini-5h` (used 25, 18000s), `gemini-weekly` (10, 604800s), `3p-5h` (0) and `3p-weekly` (50), all with scope `model` and no `3p-disabled`.
- `antigravity-denied.json` stores `usage request returned status 401`, and its stub log shows exactly one request, so a custom `base_url` never falls back.
- `antigravity-public.json` goes to Google with a fake token and stores `usage request returned status 401` from the daily host, with no prod retry.
- Each stub log line has body `{"project":"fake-project-<name>"}` and an `antigravity/hub/...` user agent.

The prod fallback after a daily 5xx or an unreadable 200 can't be driven live, because the public hosts can't be stubbed. `internal/quotahistory/recorder_test.go` covers it.

## Gotchas

- Fake tokens are rejected by Anthropic and Google, so a healthy fake account with no stub still stores an error reading with no windows. That is the expected result, not a failure.
- There is no on-demand poll endpoint. The first poll is at +30s, then every 5 minutes, so anything that waits on a poll costs minutes.
- The recorder reads the account's current access token and never refreshes it. An expired token is recorded as `access token expired` and the refresh loop handles the account separately.
- Files live next to the config file, not in the auth dir: `/tmp/verify-clepsydra-proxy/run/quota-history/` for the isolated instance.
- `since` older than the 21-day retention is clamped to the retention cutoff, and a malformed value is a 400 rather than a default.
- `launch.sh` logs `quota history recorder started (interval=5m0s, dir=...)` at startup, which proves the recorder is running before any poll lands.

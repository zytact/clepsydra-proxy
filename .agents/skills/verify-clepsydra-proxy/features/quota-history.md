# Quota history

The background recorder that reads every enabled Claude and Codex account's usage windows on a timer and keeps a JSON line per account per poll on disk.

## Sub-features

- The recorder starts 30 seconds after the proxy and polls every 5 minutes.
- Each poll appends one JSON line per enabled Claude or Codex account to `quota-history/<UTC date>.jsonl`, in the folder that holds the config file.
- A failed poll stores the reason in `error` and no windows, so a rejected or expired token still leaves a reading.
- Disabled accounts are skipped, and so is every other provider.
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

## Gotchas

- Fake tokens are rejected by Anthropic, so a healthy fake account still stores an error reading with no windows. That is the expected result, not a failure.
- There is no on-demand poll endpoint. The first poll is at +30s, then every 5 minutes, so anything that waits on a poll costs minutes.
- The recorder reads the account's current access token and never refreshes it. An expired token is recorded as `access token expired` and the refresh loop handles the account separately.
- Files live next to the config file, not in the auth dir: `/tmp/verify-clepsydra-proxy/run/quota-history/` for the isolated instance.
- `since` older than the 21-day retention is clamped to the retention cutoff, and a malformed value is a 400 rather than a default.
- `launch.sh` logs `quota history recorder started (interval=5m0s, dir=...)` at startup, which proves the recorder is running before any poll lands.

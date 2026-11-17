# Token refresh

The auto-refresh loop renews OAuth access tokens before or after they expire, and backs off after failures.

## Sub-features

- An expired token is refreshed as soon as the account loads.
- A rejected refresh token (`invalid_grant`) backs off and marks the account `invalid grant (retrying)`.
- A refresh that fails before connecting retries after 15 seconds, then 30, 60, 120 and 240 seconds, capped at 5 minutes. A failure before connecting is a DNS failure, a refused dial, or a `proxy-url` whose proxy is not up yet.
- Any other refresh failure, including a 5xx from the token endpoint, keeps the flat 5-minute retry.

## How to get to it (user POV)

- Start the proxy with accounts whose tokens expired while it was down.
- Start with a `proxy-url` that is not up yet; the account recovers once the proxy answers.
- `POST /v8/management/credentials/refresh` with `{"names":["<file>"]}` from the panel's refresh button.

## Driving it with api.sh

```bash
$P/fake-claude.sh beta --expired
grep 'Token refresh attempt' /tmp/verify-clepsydra-proxy/run/auths/logs/main.log
$P/api.sh GET /credentials
```

Proof: `main.log` shows `Token refresh attempt 1 failed: token refresh failed with status 400: {"error": "invalid_grant"...` for the fake token, and `/credentials` reports `status: error`, `status_message: invalid grant (retrying)`.

Driving the network schedule needs an account whose dials fail:

```bash
$P/fake-claude.sh gamma --expired --offline
# watch main.log for about 2.5 minutes
grep 'connection refused' /tmp/verify-clepsydra-proxy/run/auths/logs/main.log
```

Proof: attempts land at roughly +0s, +15s, +45s and +105s from the account loading, so the gaps are 15s, 30s and 60s. Removing the `proxy_url` line from the auth file while keeping the token expired restores the network for the account, and the next attempt reaches Anthropic and fails with `invalid_grant` instead of a dial error. That attempt can be up to 4 minutes out, because the last offline failure scheduled a 240-second delay. `/credentials` moves from `status_message: token expired` to `invalid grant (retrying)`.

## Gotchas

- The fake refresh token reaches Anthropic's real token endpoint and is rejected there. Use `--offline` to keep the attempt local.
- `--offline` sets a per-account `proxy_url` to a closed port, so dials fail with `connection refused`, a `net.OpError` like the DNS failure at boot.
- `Token refresh attempt` lines carry no account id. Tell the offline attempts apart by the closed-port address in the dial error, and expect `invalid_grant` lines from other expired fake accounts to interleave.
- The retry schedule lives with the account in memory. Rewriting the auth file with the same tokens does not reset a pending retry, so after an attempt that scheduled 4 minutes the fixed version waits that out.
- The first attempt happens at once, so its log line sits in the same second as the `auth file changed (CREATE)` line for the account. Measure the first gap from the attempt itself, not from the file write.

# Token refresh

The auto-refresh loop renews OAuth access tokens before or after they expire, and backs off after failures.

## Sub-features

- An expired token is refreshed as soon as the account loads.
- A rejected refresh token (`invalid_grant`) backs off and marks the account `invalid grant (retrying)`.
- A refresh that fails because the network is unreachable retries on its own schedule.

## How to get to it (user POV)

- Start the proxy with accounts whose tokens expired while it was down.
- `POST /v8/management/credentials/refresh` with `{"names":["<file>"]}` from the panel's refresh button.

## Driving it with api.sh

```bash
$P/fake-claude.sh beta --expired
grep 'Token refresh attempt' /tmp/verify-clepsydra-proxy/run/auths/logs/main.log
$P/api.sh GET /credentials
```

Proof: `main.log` shows `Token refresh attempt 1 failed: token refresh failed with status 400: {"error": "invalid_grant"...` for the fake token, and `/credentials` reports `status: error`, `status_message: invalid grant (retrying)`.

## Gotchas

- The fake refresh token reaches Anthropic's real token endpoint and is rejected there. Use `--offline` to keep the attempt local.
- `--offline` sets a per-account `proxy_url` to a closed port, so dials fail with `connection refused`, a `net.OpError` like the DNS failure at boot.

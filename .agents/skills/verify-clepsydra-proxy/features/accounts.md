# Accounts

The OAuth credentials the proxy loads from its auth dir and serves through the Management API.

## Sub-features

- A file added to the auth dir appears in `/credentials` without a restart.
- A deleted file disappears.
- Each entry carries `auth_index`, `provider`, `status`, `status_message` and `weight`.

## How to get to it (user POV)

- Sign in with `--claude-login`, or drop an auth file into the auth dir.
- The panel's Auth Files page reads `GET /v8/management/credentials`.

## Driving it with api.sh

```bash
$P/fake-claude.sh alpha
$P/api.sh GET /credentials
rm /tmp/verify-clepsydra-proxy/run/auths/claude-alpha.json
$P/api.sh GET /credentials
```

Proof: `claude-alpha.json` is listed with `provider: claude` after the add and gone after the delete, and `main.log` shows `auth file changed (CREATE)` and `(REMOVE)`.

## Gotchas

- A fake account with a valid-looking token is `active` until something calls upstream with it.

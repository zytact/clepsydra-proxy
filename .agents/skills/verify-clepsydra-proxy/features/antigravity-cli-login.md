# Antigravity CLI login

The proxy reads the Antigravity CLI's own Google login and shows it as a quota-only Antigravity credential. Requests never run on it.

## Sub-features

- The reader takes the Windows Credential Manager entry `gemini:antigravity` on Windows, and `~/.gemini/antigravity-cli/antigravity-oauth-token` everywhere else. It never writes back.
- A loop reads the login once at startup and then every 60 seconds. It adds `antigravity-cli` when a login appears, and replaces it when the refresh token changes. It removes the credential when the file disappears or can't be parsed, for example when there's no refresh token. It logs `antigravity cli login changed state` once per state change: `absent`, `unreadable`, `disabled`, or an error when project discovery fails.
- The credential lives only in memory. Nothing is written to the auth dir, and the watcher never sees it.
- `GET /credentials` lists it with `id`/`name` `antigravity-cli`, `provider: antigravity`, `quota_only: true`, `login_source: "antigravity-cli"`, `runtime_only: true`, `source: "memory"`, its `email`, and `project_id` once `loadCodeAssist` finds one.
- Download, `DELETE /credentials`, `PATCH /credentials/status` and `PATCH /credentials/fields` return 409. `POST /credentials/verify` also returns 409, with `auth_quota_only`.
- `POST /requests/api-call` with its `auth_index` still works. This is how the panel reads its quota.
- It registers no models, so `/v1/models` doesn't grow and a model request can't land on it. Every executor call goes through `sdk/cliproxy/auth/execution_guard.go`, which refuses quota-only credentials.
- `oauth.providers.antigravity.cli-login-quota: false` turns the reader off and removes the credential on the next tick.

## How to get to it (user POV)

- Sign in to the Antigravity CLI on the machine that runs the proxy. The credential appears within a minute.
- The panel's Auth Files and Quota pages show it as "From Antigravity CLI".

## Driving it with a fake login

`launch.sh` runs the proxy with `HOME=$STATE_DIR/home`, so it never sees the real `~/.gemini` login. `fake-cli-login.sh [--corrupt|--remove]` writes, breaks or deletes a fake login under that home.

```bash
$P/fake-cli-login.sh
# within 60 seconds
$P/api.sh GET /credentials
$P/api.sh GET '/credentials/download?name=antigravity-cli' -w ' [%{http_code}]'
$P/api.sh DELETE '/credentials?name=antigravity-cli' -w ' [%{http_code}]'
$P/api.sh PATCH /credentials/status -H 'Content-Type: application/json' -d '{"name":"antigravity-cli","disabled":true}' -w ' [%{http_code}]'
$P/api.sh PATCH /credentials/fields -H 'Content-Type: application/json' -d '{"name":"antigravity-cli","prefix":"x"}' -w ' [%{http_code}]'
$P/api.sh POST /credentials/verify -H 'Content-Type: application/json' -d '{"name":"antigravity-cli"}' -w ' [%{http_code}]'
ls "$AUTH_DIR"
source /tmp/verify-clepsydra-proxy/session.env
curl -s -H 'Authorization: Bearer verify-client-key' "$BASE/v1/models"
curl -s -H 'Authorization: Bearer verify-client-key' -H 'Content-Type: application/json' "$BASE/v1/chat/completions" \
  -d '{"model":"gemini-3-pro-high","messages":[{"role":"user","content":"hi"}]}'
$P/fake-cli-login.sh --corrupt   # gone within 60 seconds
$P/fake-cli-login.sh             # back within 60 seconds
$P/fake-cli-login.sh --remove    # gone within 60 seconds
grep 'cli login changed' "$RUN_DIR/auths/logs/main.log"
```

Proof:
- The list entry carries the fields above.
- All five mutations answer 409.
- `$AUTH_DIR` holds no `antigravity-cli` file.
- `/v1/models` returns `{"data":[]}`, and the chat request is a 400 `unknown provider for model`.
- After `--corrupt` and `--remove`, the credential leaves the list within a minute, and the log shows `unreadable` and `absent`.

For `api-call`, point it at a local listener (the PR 1 branch's `quota-stub.py`, or any HTTP server) with the credential's `auth_index` and `"header":{"Authorization":"Bearer $TOKEN$"}`. Proof is `status_code: 200` and the listener seeing the request.

For the config key, write `cli-login-quota: false` under `oauth.providers.antigravity` in `$RUN_DIR/config.yaml`, then write it back to `true`. Proof is `state=disabled` in the log, the credential gone within a minute, and back a minute after the key returns to `true`.

The credits fallback and pinned-auth refusals have no HTTP entry point. `sdk/cliproxy/auth/quota_only_test.go`, `execution_guard_test.go` and `execution_guard_source_test.go` cover them.

## Gotchas

- Project discovery calls `loadCodeAssist` with the fake token. Google answers 401, so the log shows `antigravity auth missing project_id` and the entry has no `project_id`. That is expected, and the credential is still listed.
- Edit the config in place (`python3 -c "open(p,'w').write(...)"`), not with `sed -i`. In one run a `sed -i` rename made the config watcher stop reloading until relaunch.
- A read-only quota check against the user's real Antigravity login needs the user's explicit approval each time. Never copy the real token file or Credential Manager entry into the isolated home.

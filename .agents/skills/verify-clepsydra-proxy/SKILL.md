---
name: verify-clepsydra-proxy
description: Build and run an isolated Clepsydra proxy (this repo, a CLIProxyAPI fork) on its own port, config and auth dir, then drive its Management API with curl and read its logs. Use to prove a change to the proxy's routes, background loops or token refresh, without touching the user's real proxies or accounts.
---

# Verify clepsydra-proxy

The proxy is a Go HTTP server. Verification builds this checkout, runs it on `127.0.0.1:18317` with a generated config, an empty auth dir and its own random management key, and drives `/v8/management/*` with curl.

**Never point this at the user's real accounts.** Do not copy files from `~/.cli-proxy-api` or `/mnt/c/Users/*/.cli-proxy-api`: two proxies refreshing the same OAuth tokens break both. Use `fake-claude.sh` for accounts. Its tokens are placeholders, so upstream rejects every call, which is enough to drive refresh and polling paths for real.

The user's live proxies run released binaries; `~/dotfiles/AGENTS.md` says where they run and how they are reached. Swapping their binary needs the user's explicit approval each time. Read-only calls to them use the user's management key; a proxy bans an address after five wrong keys, so never guess one.

All commands run from the repo root. `$P` is `.agents/skills/verify-clepsydra-proxy/scripts`.

## Launch

```bash
P=.agents/skills/verify-clepsydra-proxy/scripts
$P/launch.sh
```

It runs `go build ./cmd/server`, writes `/tmp/verify-clepsydra-proxy/run/config.yaml`, starts the binary with that config, and waits until `GET /v8/management/config` answers 200. Ready when it prints `ready: http://127.0.0.1:18317 ...`. `VERIFY_PORT` and `VERIFY_STATE_DIR` override the port and state dir, so two runs can coexist. It refuses to start while `session.env` exists.

Run it directly, not through a pipe such as `| tail`: the pipe stays open after the script returns and the shell hangs.

## Doctor

```bash
source /tmp/verify-clepsydra-proxy/session.env && echo "$BASE"
$P/api.sh GET /config | head -c 200
head -3 /tmp/verify-clepsydra-proxy/proxy.out
```

Healthy: a base URL, a JSON config whose `server.port` is 18317, and the startup lines `CLIProxyAPI Version: dev` and `API server started successfully on: 127.0.0.1:18317`.

## Drive

`api.sh <METHOD> <path under /v8/management> [curl args...]` calls the instance with its own key. `fake-claude.sh <name> [--expired] [--offline]` writes a fake Claude account: `--expired` sets the access token to have expired an hour ago, and `--offline` routes the account through a closed local port so every upstream dial fails, as it does before DNS is up.

```bash
$P/api.sh GET /credentials                      # accounts the instance holds
$P/fake-claude.sh alpha                          # fake Claude account, token valid for 8 hours
$P/fake-claude.sh beta --expired                 # expired token: the refresh loop picks it up at once
$P/fake-claude.sh gamma --expired --offline      # every upstream dial fails, as before DNS is up
rm /tmp/verify-clepsydra-proxy/run/auths/claude-gamma.json   # the watcher drops the account
```

The watcher picks up auth file changes within a second. Logs:

- `/tmp/verify-clepsydra-proxy/run/auths/logs/main.log`: the structured log, with refresh attempts and background loops.
- `/tmp/verify-clepsydra-proxy/proxy.out`: stdout and stderr, including startup.
- `/tmp/verify-clepsydra-proxy/run/`: files the instance writes next to its config.

The feature map in [features/README.md](features/README.md) lists every feature, its entry points and its proof.

## Evidence

```bash
EVIDENCE=~/.cache/verify-clepsydra-proxy/evidence/<branch-or-task>
mkdir -p "$EVIDENCE"
```

- Save the exact responses (`$P/api.sh ... > "$EVIDENCE/<name>.json"`) and the log lines that show the behavior (`grep ... main.log > "$EVIDENCE/<name>.log"`), with timestamps.
- Prove timing with timestamps from the log, not with sleeps in the script.
- Prove side effects on disk by listing the files the instance wrote under `run/`.
- Fake accounts carry `@example.invalid` addresses, so this evidence is safe to attach to a PR.

## Cleanup

```bash
.agents/skills/verify-clepsydra-proxy/scripts/cleanup.sh
```

It kills the process group `launch.sh` recorded and removes `run/`, `session.env` and `proxy.out`. The built binary in `bin/` stays for the next run, and evidence under `~/.cache/verify-clepsydra-proxy/evidence/` stays.

# Clepsydra proxy

A small fork of [CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI) that serves the [Clepsydra](https://github.com/zytact/clepsydra) management panel.

Everything CLIProxyAPI does, this does the same way. Request handling is untouched. Read the [upstream README](https://github.com/router-for-me/CLIProxyAPI#readme) for setup, providers, configuration and the SDK.

## Set up

1. Download `clepsydra-proxy_<version>_windows_amd64.zip` or `_linux_amd64.tar.gz` from [Releases](https://github.com/zytact/clepsydra-proxy/releases) and unpack it. Other platforms build from source with `go build -o cli-proxy-api ./cmd/server`.
2. Copy `config.example.yaml` to `config.yaml` next to the binary and edit two things:
   - `management.secret-key`: a password for the panel. The proxy hashes it on first start. Set `management.allow-remote: true` only if you will open the panel from another machine.
   - `access.api-keys`: replace the `your-api-key-*` placeholders with your own keys. Clients send one of these to use the proxy, and the proxy listens on every interface by default.
3. Sign in your accounts, for example `./cli-proxy-api --claude-login` or `--codex-login`. Each login writes one auth file into `oauth.auth-dir` (`~/.cli-proxy-api` by default). On a machine without a browser add `--no-browser` and see the [upstream README](https://github.com/router-for-me/CLIProxyAPI#readme).
4. From that folder, run `./cli-proxy-api` (`cli-proxy-api.exe` on Windows), or pass `--config <path>`. It listens on `port` (8317 by default).
5. Open `http://<host>:8317/management.html` and log in with the secret key. The proxy downloads the [Clepsydra](https://github.com/zytact/clepsydra) panel from its latest release on first visit and checks its sha256 digest.

Quota history starts filling 30 seconds after start. The panel's quota page needs a day of it before "Usual" and the drain band have anything to show.

## What differs from upstream

- The management panel it downloads comes from `zytact/clepsydra` releases, and the version check reads `zytact/clepsydra-proxy` releases. `management.panel-github-repository` still overrides the panel source, so a config that names the upstream panel there keeps getting it. Clear the key or set it to `https://github.com/zytact/clepsydra`. Upstream's unverified fallback page is off: with no release to download, the proxy serves no panel rather than one it cannot check.
- Release archives are named `clepsydra-proxy_<version>_<os>_<arch>`. The binary inside is still `cli-proxy-api`, so existing service definitions keep working.
- Releases ship two archives, `windows_amd64` and `linux_amd64`.
- When a token refresh fails because the network is down, such as `lookup platform.claude.com: no such host` right after boot, an expired token retries after 15 seconds instead of 5 minutes, doubling while the network stays down until it reaches the usual 5 minutes. It works again soon after the network returns. A local `proxy-url` that has not started yet counts as the network being down.
- A quota recorder reads every enabled Claude and Codex account's usage windows every 5 minutes and keeps 21 days of readings. See [Quota history](#quota-history).

## Quota history

The recorder starts 30 seconds after the proxy and polls every 5 minutes. Each poll appends one JSON line per account to `quota-history/<UTC date>.jsonl`, in the folder that holds the config file. Day files older than 21 days are deleted. It reads with each account's current access token and never refreshes one. A poll that fails stores the reason, such as `access token expired` or `usage request returned status 429`, and no windows.

`GET /v8/management/quota/history?since=<RFC 3339 time>` returns the stored readings oldest first. `since` defaults to 15 days ago.

```json
{
  "interval_seconds": 300,
  "since": "2026-11-02T12:00:00Z",
  "readings": [
    {
      "at": "2026-11-17T12:00:00Z",
      "provider": "claude",
      "auth_id": "claude-example.json",
      "auth_index": "62cb39bfdee58cf6",
      "email": "user@example.com",
      "windows": [
        { "id": "five-hour", "used_percent": 7, "resets_at": "2026-11-17T17:00:00Z", "window_seconds": 18000, "scope": "all" },
        { "id": "seven-day", "used_percent": 100, "resets_at": "2026-11-18T02:00:00Z", "window_seconds": 604800, "scope": "all" }
      ]
    }
  ]
}
```

Window ids match the management panel's: Claude `five-hour`, `seven-day`, `seven-day-fable`, `seven-day-opus`, `seven-day-sonnet`, `seven-day-oauth-apps`, `seven-day-cowork`; Codex `five-hour`, `weekly`, `monthly`. A window with scope `all` draws from every model; `model` draws from some.

## Releases

Push a tag shaped like `v<upstream version>-clepsydra.<n>`, for example `v8.0.76-clepsydra.1`. The `release` workflow builds both archives, attaches them with `checksums.txt`, and writes the notes. Other tags, including upstream's own, publish nothing.

## Staying close to upstream

The Go module path and the package names are upstream's. New features live in new files, and edits to upstream files stay to a few lines, so merging `upstream/main` stays cheap:

```bash
git fetch upstream
git merge upstream/main
```

The upstream-only workflows `docker-image`, `auto-retarget-main-pr-to-dev` and `agents-md-guard` are disabled in this repository's Actions settings, not deleted, so their files still merge cleanly.

## License

MIT, as upstream. See [LICENSE](LICENSE).

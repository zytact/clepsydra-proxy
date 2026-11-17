# Clepsydra proxy

A small fork of [CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI) that serves the [Clepsydra](https://github.com/zytact/clepsydra) management panel.

Everything CLIProxyAPI does, this does the same way. Request handling is untouched. Read the [upstream README](https://github.com/router-for-me/CLIProxyAPI#readme) for setup, providers, configuration and the SDK.

## Set up

1. Download `clepsydra-proxy_<version>_<os>_amd64` from [Releases](https://github.com/zytact/clepsydra-proxy/releases) and unpack it.
2. Copy `config.example.yaml` to `config.yaml` next to the binary. Set `management.secret-key` to a password of your choice. The proxy hashes it on first start. Set `management.allow-remote: true` only if you will open the panel from another machine.
3. Sign in your accounts, for example `./cli-proxy-api --claude-login` or `--codex-login`. Each login writes one auth file into `auth-dir`.
4. Run `./cli-proxy-api` (`cli-proxy-api.exe` on Windows). It listens on `port` (8317 by default).
5. Open `http://<host>:8317/management.html` and log in with the secret key. The proxy downloads the [Clepsydra](https://github.com/zytact/clepsydra) panel from its latest release on first visit and checks its sha256 digest.

Quota history starts filling 30 seconds after start. The panel's quota page needs a day of it before "Usual" and the drain band have anything to show.

## What differs from upstream

- The management panel it downloads comes from `zytact/clepsydra` releases, and the version check reads `zytact/clepsydra-proxy` releases. `management.panel-github-repository` still overrides the panel source, so a config that names the upstream panel there keeps getting it. Clear the key or set it to `https://github.com/zytact/clepsydra`. Upstream's unverified fallback page is off: with no release to download, the proxy serves no panel rather than one it cannot check.
- Release archives are named `clepsydra-proxy_<version>_<os>_<arch>`. The binary inside is still `cli-proxy-api`, so existing service definitions keep working.
- Releases ship two archives, `windows_amd64` and `linux_amd64`.

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

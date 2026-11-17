# Release sources

Where the proxy looks for its own new versions and for the management panel it serves.

## Sub-features

- `GET /v8/management/server/latest-version` reads the latest release of `zytact/clepsydra-proxy`.
- The panel download reads the latest release of `zytact/clepsydra` unless `management.panel-github-repository` overrides it. The legacy layout calls that section `remote-management`.
- Parsing fills the panel default in, so an absent key still means `zytact/clepsydra`; the updater falls back to the same fork when the key is empty or unparseable.

## How to get to it (user POV)

- The panel's System page shows the latest proxy version.
- The proxy downloads `management.html` at startup and on a timer when the control panel is enabled.

## Driving it with api.sh

```bash
$P/api.sh GET /server/latest-version
```

Proof: the response is `{"latest-version":"<tag>"}` naming a `zytact/clepsydra-proxy` release, or, when the fork has no release, a 502 `unexpected_status` carrying GitHub's 404 for that repo. Never a `router-for-me` version.

## Gotchas

- `GET /config` serves the persisted v8 document, so `management.panel-github-repository` is absent from a config that never sets it, even though parsing fills in `https://github.com/zytact/clepsydra`. Prove the default from `internal/config/config_defaults.go` and `internal/managementasset/updater.go`, or set the key explicitly to prove the override.
- `launch.sh` sets `disable-control-panel: true`, so the instance never downloads the panel. Drive the panel download with a separate config only when the change touches it.
- Neither fork repo has a published release right now, so a control-panel-enabled run logs a release fetch failure instead of downloading a panel.

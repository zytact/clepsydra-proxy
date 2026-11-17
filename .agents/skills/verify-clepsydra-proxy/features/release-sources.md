# Release sources

Where the proxy looks for its own new versions and for the management panel it serves.

## Sub-features

- `GET /v8/management/server/latest-version` reads the latest release of `zytact/clepsydra-proxy`.
- The panel download reads the latest release of `zytact/clepsydra` unless `management.panel-github-repository` overrides it.

## How to get to it (user POV)

- The panel's System page shows the latest proxy version.
- The proxy downloads `management.html` at startup and on a timer when the control panel is enabled.

## Driving it with api.sh

```bash
$P/api.sh GET /server/latest-version
```

Proof: the response names a `zytact/clepsydra-proxy` release, or reports that none exists yet, and never a `router-for-me` one.

## Gotchas

- `launch.sh` sets `disable-control-panel: true`, so the instance never downloads the panel. Drive the panel download with a separate config only when the change touches it.

# Feature map

One file per user-facing behavior of clepsydra-proxy that verification drives. Each says how to reach it, how to drive it, and what proves it works.

| Feature | File | Entry points |
| --- | --- | --- |
| Accounts | [accounts.md](accounts.md) | `GET /v8/management/credentials`, files in the auth dir |
| Token refresh | [token-refresh.md](token-refresh.md) | The auto-refresh loop, `POST /v8/management/credentials/refresh` |
| Quota history | [quota-history.md](quota-history.md) | `GET /v8/management/quota/history`, files in `quota-history/` next to the config |
| Release sources | [release-sources.md](release-sources.md) | `GET /v8/management/server/latest-version`, the management panel download |
| Antigravity CLI login | [antigravity-cli-login.md](antigravity-cli-login.md) | The CLI login file or Credential Manager entry, `GET /v8/management/credentials`, the credential mutations, `POST /v8/management/requests/api-call` |

A proof covers every entry point its feature lists.

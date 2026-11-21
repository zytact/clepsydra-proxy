# Changelog

## [8.1.0-clepsydra.1](https://github.com/zytact/clepsydra-proxy/compare/v8.0.76-clepsydra.2...v8.1.0-clepsydra.1) (2026-11-21)


### ⚠ BREAKING CHANGES

* **copilot:** upstream-only Copilot models without a registry preset are no longer routed, and the persisted copilot_catalog metadata, ReadCatalog and string FetchCatalog APIs are removed

### Features

* **antigravity:** scope synthesized request.sessionId per credential ([3a24d03](https://github.com/zytact/clepsydra-proxy/commit/3a24d03cdd4381dcb0dac68e7227e8b0e6a0509e)), closes [#7706](https://github.com/zytact/clepsydra-proxy/issues/7706)
* **copilot:** add GitHub Copilot provider support ([7ae771f](https://github.com/zytact/clepsydra-proxy/commit/7ae771fcb49dc35802393df92983d4490e08f15b))
* **copilot:** add preset-based in-memory catalog ([d059cf2](https://github.com/zytact/clepsydra-proxy/commit/d059cf2891bff7350dbb11aac1cf9e04448e6b3e))
* **copilot:** support Home-dispatched Copilot execution ([1ef770d](https://github.com/zytact/clepsydra-proxy/commit/1ef770d9a1be9333cbd6b29301208057f2c6641d))
* **copilot:** support OAuth model aliases and exact route-aware selection ([5dff8b6](https://github.com/zytact/clepsydra-proxy/commit/5dff8b6728dac0d531d9254d6a687463af48d9d8))
* **management:** persist server identity and expose it to authenticated clients ([6960763](https://github.com/zytact/clepsydra-proxy/commit/6960763455c4ac3d4828fb1736aa5de2183ccecf))


### Bug Fixes

* **antigravity:** apply Host, custom auth headers, and TOS_VIOLATION handling in credits query ([acb56c6](https://github.com/zytact/clepsydra-proxy/commit/acb56c6cba602715521395aa3104105abc03f40a)), closes [#7820](https://github.com/zytact/clepsydra-proxy/issues/7820)
* **antigravity:** bound auth response reads, preserve error details, and make onboarding sleep ctx-aware ([91ccf28](https://github.com/zytact/clepsydra-proxy/commit/91ccf28400779ef9249c7873e6b747f2db61c1c7)), closes [#7759](https://github.com/zytact/clepsydra-proxy/issues/7759)
* **antigravity:** enforce tos-pause-window, Host, and custom auth headers in HttpRequest ([d88d8dc](https://github.com/zytact/clepsydra-proxy/commit/d88d8dcb5c855e8a507001187d418559fcd3a95c)), closes [#7824](https://github.com/zytact/clepsydra-proxy/issues/7824)
* **antigravity:** filter ineligible credentials and resolve aliases in credits fallback ([e0a1e93](https://github.com/zytact/clepsydra-proxy/commit/e0a1e93d2b6594e823d3248ada6ed771e881183f)), closes [#7757](https://github.com/zytact/clepsydra-proxy/issues/7757)
* **antigravity:** harden SSE JSON reassembly and mid-stream error bookkeeping ([892fc48](https://github.com/zytact/clepsydra-proxy/commit/892fc4864cc46a993c8f2e7bef0377ac91aa9c50)), closes [#7753](https://github.com/zytact/clepsydra-proxy/issues/7753)
* **antigravity:** honor explicit lower and zero max-idle-conns-per-host pool limits ([80fb55d](https://github.com/zytact/clepsydra-proxy/commit/80fb55d8ebaf9a17ed19b7278074d83e71a0214f)), closes [#7758](https://github.com/zytact/clepsydra-proxy/issues/7758)
* **antigravity:** honor HTTP Retry-After header fallback on 429 responses ([84b838e](https://github.com/zytact/clepsydra-proxy/commit/84b838e1c6525420cae7839d50e9553f2c8e21b6)), closes [#7821](https://github.com/zytact/clepsydra-proxy/issues/7821)
* **antigravity:** mark TOS pause 503 error as request-scoped ([81018ba](https://github.com/zytact/clepsydra-proxy/commit/81018ba473b2dc48ec0bf43afa1006b779cf6dac)), closes [#7748](https://github.com/zytact/clepsydra-proxy/issues/7748)
* **antigravity:** match claude 5.5 low, medium, and max effort variants for structured output ([c813f77](https://github.com/zytact/clepsydra-proxy/commit/c813f779ecac362f2c83fe05495dc9488027d700)), closes [#7823](https://github.com/zytact/clepsydra-proxy/issues/7823)
* **antigravity:** persist capability hints freshness metadata and initialize cache expiry on seed ([533edd3](https://github.com/zytact/clepsydra-proxy/commit/533edd3814e0ad48a502421d4fb93257bbaecb9b)), closes [#7818](https://github.com/zytact/clepsydra-proxy/issues/7818)
* **antigravity:** persist execution-time token refreshes and preserve refresh context values ([cdbda2f](https://github.com/zytact/clepsydra-proxy/commit/cdbda2f2dc758487e939f684ef231a82ed9dfeac)), closes [#7749](https://github.com/zytact/clepsydra-proxy/issues/7749)
* **antigravity:** preserve project discovery error classification ([204be8b](https://github.com/zytact/clepsydra-proxy/commit/204be8b9ae007fc58d63a2654290a494e827d77a)), closes [#7751](https://github.com/zytact/clepsydra-proxy/issues/7751)
* **antigravity:** preserve thoughtSignature on text parts and empty-text carriers ([9176cb2](https://github.com/zytact/clepsydra-proxy/commit/9176cb22e148dc4f31b646c8d7d362efd73fd555)), closes [#7755](https://github.com/zytact/clepsydra-proxy/issues/7755)
* **antigravity:** publish unavailable credits hint when no GOOGLE_ONE_AI entry matches ([8e5a98e](https://github.com/zytact/clepsydra-proxy/commit/8e5a98e886a6e382b63ff0752ff7aaac5a4c7159)), closes [#7750](https://github.com/zytact/clepsydra-proxy/issues/7750)
* **antigravity:** scan all user parts for session ID and match models case-insensitively ([18a03e4](https://github.com/zytact/clepsydra-proxy/commit/18a03e40c18eadb6a21f57db5a58dbb7248a783d)), closes [#7760](https://github.com/zytact/clepsydra-proxy/issues/7760)
* **antigravity:** scope 403 validation and 401 inference errors to credential ([d3e4db8](https://github.com/zytact/clepsydra-proxy/commit/d3e4db8b06036ac4898eb05fb395cb9f4091db4b)), closes [#7814](https://github.com/zytact/clepsydra-proxy/issues/7814)
* **antigravity:** separate 403 probe backoff from token rotation and skip disabled auths ([f8e297a](https://github.com/zytact/clepsydra-proxy/commit/f8e297aabf74a7723b2906ca0298b4b066a99956)), closes [#7703](https://github.com/zytact/clepsydra-proxy/issues/7703)
* **antigravity:** skip model capability probes for unauthenticated credentials ([f48975b](https://github.com/zytact/clepsydra-proxy/commit/f48975b4ddfcd2e33828cc4fa40504d6218d09ff)), closes [#7817](https://github.com/zytact/clepsydra-proxy/issues/7817)
* **antigravity:** update Hub fallback and changelist map to 2.34.0 ([093c7ef](https://github.com/zytact/clepsydra-proxy/commit/093c7efaac905e0f92f9147eb0ec7ba25921d97d)), closes [#7813](https://github.com/zytact/clepsydra-proxy/issues/7813)
* **antigravity:** wire version updater into service lifecycle and honor proxy config ([4ea1527](https://github.com/zytact/clepsydra-proxy/commit/4ea1527c948314dac587c1e46c44ac8174d6b5b3)), closes [#7822](https://github.com/zytact/clepsydra-proxy/issues/7822)
* **api:** decode gzip request bodies and honor Content-Encoding on Claude routes ([6e7a1b2](https://github.com/zytact/clepsydra-proxy/commit/6e7a1b294aabf440fe7899cd39ef02ef855c1e6e)), closes [#7711](https://github.com/zytact/clepsydra-proxy/issues/7711)
* **auth:** align executeHome and executeHomeOnce with execution-path parity ([84e0cd9](https://github.com/zytact/clepsydra-proxy/commit/84e0cd9b23f74d41974151db2cf61ca043031659)), closes [#7724](https://github.com/zytact/clepsydra-proxy/issues/7724)
* **auth:** align VerifyInference with per-auth transport, auth prep, and cancellation ([8906442](https://github.com/zytact/clepsydra-proxy/commit/8906442bd4b4c8a19319228e9ab1970dd8f2cb8f)), closes [#7737](https://github.com/zytact/clepsydra-proxy/issues/7737)
* **auth:** cancel abandoned stream attempts before draining ([538f961](https://github.com/zytact/clepsydra-proxy/commit/538f961107925fcc419105ae6a809f7aaa8d4c57)), closes [#7745](https://github.com/zytact/clepsydra-proxy/issues/7745)
* **auth:** capture per-attempt access token on streaming failure results ([f5c5681](https://github.com/zytact/clepsydra-proxy/commit/f5c5681a8361c9da53e93c60b6507816eee7468b)), closes [#7731](https://github.com/zytact/clepsydra-proxy/issues/7731)
* **auth:** check active plugin scheduler in affinity lookup and normalize mixed legacy provider keys ([221d5e0](https://github.com/zytact/clepsydra-proxy/commit/221d5e0818708fd02117e753ce1dbc26ab81f238)), closes [#7761](https://github.com/zytact/clepsydra-proxy/issues/7761)
* **auth:** classify anthropic stop_reason consistently for message and message_delta payloads ([3a0d302](https://github.com/zytact/clepsydra-proxy/commit/3a0d3023d0090e51290813bce96ebf139f586b99)), closes [#7722](https://github.com/zytact/clepsydra-proxy/issues/7722)
* **auth:** clear stream rewriter before emitting flushed tail in wrapStreamResult ([cdd6ae6](https://github.com/zytact/clepsydra-proxy/commit/cdd6ae6d9b10042c24246f8ef7ebb1a56b1da045)), closes [#7738](https://github.com/zytact/clepsydra-proxy/issues/7738)
* **auth:** clone metadata before interceptor session updates and publish refreshed auth on 401 retry ([2dbad4d](https://github.com/zytact/clepsydra-proxy/commit/2dbad4ddf0101af4c39249feec568780d2da9528)), closes [#7725](https://github.com/zytact/clepsydra-proxy/issues/7725)
* **auth:** clone per-attempt opts.Metadata from immutable base metadata ([8f11580](https://github.com/zytact/clepsydra-proxy/commit/8f11580852166369d540ab9ad1b84f4dfe429bdb)), closes [#7733](https://github.com/zytact/clepsydra-proxy/issues/7733)
* **auth:** close canceled streams during upstream stalls ([4aca8df](https://github.com/zytact/clepsydra-proxy/commit/4aca8df3d3302874a95082e2bd1a6c84808bf04c)), closes [#7741](https://github.com/zytact/clepsydra-proxy/issues/7741)
* **auth:** delete existing keys case-insensitively before applying updates in mergeRequestHeaders ([486624c](https://github.com/zytact/clepsydra-proxy/commit/486624c777b05ae3b2bf3e69bcee10fe3ca75ed7)), closes [#7742](https://github.com/zytact/clepsydra-proxy/issues/7742)
* **auth:** end home selection on empty stream and drain upstream chunks on cancel ([9be2d5f](https://github.com/zytact/clepsydra-proxy/commit/9be2d5f1db6cff7c5fd1a85691497f7e858db7ea)), closes [#7723](https://github.com/zytact/clepsydra-proxy/issues/7723)
* **auth:** enforce consistent cancellation and Retry-After cooldown handling ([3ad507b](https://github.com/zytact/clepsydra-proxy/commit/3ad507bf1d3544ab184ed1038420268d2bddad53)), closes [#7720](https://github.com/zytact/clepsydra-proxy/issues/7720)
* **auth:** honor request-scoped-errors, Retry-After, and credential scope on PrepareRequestAuth failure ([2f57a5a](https://github.com/zytact/clepsydra-proxy/commit/2f57a5a67e7b0c2ea39ec353c73b06939ca6e80d)), closes [#7726](https://github.com/zytact/clepsydra-proxy/issues/7726)
* **auth:** inspect response.candidates and additional Gemini failure finishReasons in completion flags ([2c8b32f](https://github.com/zytact/clepsydra-proxy/commit/2c8b32fd3e0bef4b4f6103428523222eb3e189dc)), closes [#7739](https://github.com/zytact/clepsydra-proxy/issues/7739)
* **auth:** preserve disabled and terminal 401 states during model reconciliation ([35afdf0](https://github.com/zytact/clepsydra-proxy/commit/35afdf07f976a63af36cc14401c26da64b3e9840)), closes [#7756](https://github.com/zytact/clepsydra-proxy/issues/7756)
* **auth:** preserve disabled unavailability and skip inference verification on late success ([29de31a](https://github.com/zytact/clepsydra-proxy/commit/29de31a89da51a7bea9a38cf247c15d037acfdc5)), closes [#7754](https://github.com/zytact/clepsydra-proxy/issues/7754)
* **auth:** preserve interceptor session clearing and metadata isolation ([0147c80](https://github.com/zytact/clepsydra-proxy/commit/0147c80fe80e3bab4b03449f1af109c5b5334d84)), closes [#7735](https://github.com/zytact/clepsydra-proxy/issues/7735)
* **auth:** preserve stream headers and honor terminal request stops ([df6c8cc](https://github.com/zytact/clepsydra-proxy/commit/df6c8ccb96ee17611809d8d742fd5dd54f9aac53)), closes [#7740](https://github.com/zytact/clepsydra-proxy/issues/7740)
* **auth:** preserve streamResult.Headers on invalid-request bootstrap errors and avoid double-wrapping streamBootstrapError ([d52047a](https://github.com/zytact/clepsydra-proxy/commit/d52047a66270cd5aabb14b8aabfe8ca0db4c7b61)), closes [#7721](https://github.com/zytact/clepsydra-proxy/issues/7721)
* **auth:** publish selected auth metadata only after executable model check ([0399de3](https://github.com/zytact/clepsydra-proxy/commit/0399de3801ad894d4392a85caaaed738c2e52410)), closes [#7730](https://github.com/zytact/clepsydra-proxy/issues/7730)
* **auth:** scope stream bootstrap retry-after annotation to the same upstream attempt ([d4f6edb](https://github.com/zytact/clepsydra-proxy/commit/d4f6edb162b23ff621d02b8ef368246ff8faaf52)), closes [#7728](https://github.com/zytact/clepsydra-proxy/issues/7728)
* **auth:** serialize request auth preparation with refresh locks and abort unregistered credentials ([0a34244](https://github.com/zytact/clepsydra-proxy/commit/0a34244e32e129d43cc6ffad7b0d5cc78471296a)), closes [#7727](https://github.com/zytact/clepsydra-proxy/issues/7727)
* **auth:** stop credential loop on cancellation during auth preparation and Home execution ([dfdb3ba](https://github.com/zytact/clepsydra-proxy/commit/dfdb3ba4ee5326e181c67d7fd8da3ee23bc9b73b)), closes [#7716](https://github.com/zytact/clepsydra-proxy/issues/7716)
* **auth:** treat zero-payload streams before EOF as empty_stream ([ab97187](https://github.com/zytact/clepsydra-proxy/commit/ab971876bf708d0686a423a33923cd94ba675d6a)), closes [#7732](https://github.com/zytact/clepsydra-proxy/issues/7732)
* **auth:** use syncPreferredUpstreamError in stream failover paths ([b4247e1](https://github.com/zytact/clepsydra-proxy/commit/b4247e1918d78af3b42352375ff771dc04d0e7f8)), closes [#7734](https://github.com/zytact/clepsydra-proxy/issues/7734)
* **claude:** preserve native haiku helpers on custom base URLs ([7a5dee5](https://github.com/zytact/clepsydra-proxy/commit/7a5dee5207148ab3a97bf803dd0445c3bdbd85ed)), closes [#7809](https://github.com/zytact/clepsydra-proxy/issues/7809)
* **codex:** sync temp file before close and rename in WriteJSONFileAtomic ([6d703fd](https://github.com/zytact/clepsydra-proxy/commit/6d703fdf4818cc7d24cd23109c656cae2b2e1e25)), closes [#7718](https://github.com/zytact/clepsydra-proxy/issues/7718)
* **copilot:** return gemini format for count tokens ([b824c05](https://github.com/zytact/clepsydra-proxy/commit/b824c052975da29f37353d7402ebe4b77471d8b3))
* **copilot:** translate copilot streaming responses ([9792ac2](https://github.com/zytact/clepsydra-proxy/commit/9792ac281ea3af82b1c38f3bc5f3f05ce869cf84))
* **copilot:** use execution mode for claude stream translation ([2021118](https://github.com/zytact/clepsydra-proxy/commit/20211183c5ebd2929571406113709b0e933fbc2e))
* **devin:** preserve gpt reasoning across responses tool calls ([cb5f62f](https://github.com/zytact/clepsydra-proxy/commit/cb5f62f7f4ad5dc92dcfb68f0ab50c871e39981c)), closes [#7805](https://github.com/zytact/clepsydra-proxy/issues/7805)
* **executor:** ignore Claude Code attribution block in reasoning replay lane ([bba22ed](https://github.com/zytact/clepsydra-proxy/commit/bba22ed9becf5a8ae608822306c9363f7cde4d32)), closes [#7766](https://github.com/zytact/clepsydra-proxy/issues/7766)
* **gemini:** cap Gemini 4 max completion tokens at 65536 ([c1040cc](https://github.com/zytact/clepsydra-proxy/commit/c1040cc0d98a08be1a4f11423c9fe77b5886ee50)), closes [#7707](https://github.com/zytact/clepsydra-proxy/issues/7707)
* **meta:** synthesize unstreamed Responses messages and translate refusal parts ([ca45971](https://github.com/zytact/clepsydra-proxy/commit/ca459712c0a787cac90fd65406c5e42c7d8e540a))
* **meta:** tighten Responses output synthesis tracking and refusal fallbacks ([e442c32](https://github.com/zytact/clepsydra-proxy/commit/e442c32d4f82f38541d003561943993b9b8e6d1f))
* **proxyutil:** honor context cancellation on SOCKS5 HTTP transports ([cec8a7f](https://github.com/zytact/clepsydra-proxy/commit/cec8a7f390038f83f1507fadbe494dd32ac824ab)), closes [#7819](https://github.com/zytact/clepsydra-proxy/issues/7819)
* **translator:** correct antigravity gemini stream output ([a3e7542](https://github.com/zytact/clepsydra-proxy/commit/a3e754213ac2d24a41f80dc7f2a4005a159e0fc4)), closes [#7799](https://github.com/zytact/clepsydra-proxy/issues/7799)

## [8.0.76-clepsydra.2](https://github.com/zytact/clepsydra-proxy/compare/v8.0.76-clepsydra.1...v8.0.76-clepsydra.2) (2026-11-19)


### Features

* **antigravity:** read the CLI's own Google login ([e6a6016](https://github.com/zytact/clepsydra-proxy/commit/e6a60169f48a0ec431472f79e086da22f35e7047))
* **antigravity:** show the Antigravity CLI's quota without ever routing through it ([84b1cce](https://github.com/zytact/clepsydra-proxy/commit/84b1ccee682e2d0bbb1c1ebd7802a64050d483e3))
* **antigravity:** show the CLI login as a quota-only credential ([61a624f](https://github.com/zytact/clepsydra-proxy/commit/61a624f1bfe096eb9c88b95f62e24def06452fb3))
* **auth:** keep quota-only credentials out of routing and management edits ([e66c34b](https://github.com/zytact/clepsydra-proxy/commit/e66c34bc19d2a2e82ab3531b176834659b5dc0c1))
* **quota-history:** chart Antigravity limits over time like Claude and Codex ([c29d7fb](https://github.com/zytact/clepsydra-proxy/commit/c29d7fb153d21dbfee56cee7c6cb16cdc7802250))
* **quota-history:** record antigravity quota windows ([7fef9f1](https://github.com/zytact/clepsydra-proxy/commit/7fef9f1a38fbc8fc18670eca16968b5d47e219a2))


### Bug Fixes

* **antigravity:** discover the quota-only project without onboarding ([ca80bc5](https://github.com/zytact/clepsydra-proxy/commit/ca80bc545a1af03f3dc8d5f107a82971673f7bf7))
* **antigravity:** drop the CLI login as soon as config turns it off ([20aaf4f](https://github.com/zytact/clepsydra-proxy/commit/20aaf4fd34249915cd08452c783219566774e345))
* **antigravity:** drop the CLI login when it cannot be read ([f038ccb](https://github.com/zytact/clepsydra-proxy/commit/f038ccbfa72e9908e426b7e4fb82d9b55fee76da))
* **antigravity:** keep the CLI login out of the watcher's runtime auths ([9aad5fb](https://github.com/zytact/clepsydra-proxy/commit/9aad5fb1426fe07dc51912cb3fed887d4cf58860))
* **antigravity:** wire the CLI login only into the real server ([7cb7e21](https://github.com/zytact/clepsydra-proxy/commit/7cb7e218f1f5667ff49297b2a5e741fa1cbd9309))
* **auth:** refuse quota-only credentials at every executor call ([bb5ea3f](https://github.com/zytact/clepsydra-proxy/commit/bb5ea3f0ea2b7124e5445339b9a60645ecd73258))
* **auth:** refuse quota-only credentials when injecting request credentials ([5932d44](https://github.com/zytact/clepsydra-proxy/commit/5932d447aa74273d90bc37fff65ac0832cb56bd8))
* **management:** block Host overrides and redirects on quota-only api-calls ([b3f0b32](https://github.com/zytact/clepsydra-proxy/commit/b3f0b32d39117805ee274ed05630e0669cb5b97d))
* **management:** limit quota-only api-calls to Antigravity quota endpoints ([ab5c5af](https://github.com/zytact/clepsydra-proxy/commit/ab5c5afc60c61bfb516f4ba9b30a63dda50d3f1c))
* **management:** refuse targeted re-login of quota-only credentials ([64a650d](https://github.com/zytact/clepsydra-proxy/commit/64a650d551acb8d3e94f008cfe2aaa442f971f33))

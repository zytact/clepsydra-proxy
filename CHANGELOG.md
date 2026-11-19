# Changelog

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

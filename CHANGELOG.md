# Changelog

## [1.0.1](https://github.com/regask/backstage-cli/compare/v1.0.0...v1.0.1) (2026-07-24)


### Bug Fixes

* **contracts:** align matrix/overlays/ticket-lookup + whoami with live backend shapes ([328843f](https://github.com/regask/backstage-cli/commit/328843ff5d28d0ff84a8003145bf586f5b32bd68))
* wire --version, back off scaffolder poll loop, login timeout, unpredictable CSRF state ([4992ed9](https://github.com/regask/backstage-cli/commit/4992ed978ab4937a5a43fce5db7efd59b1e58b4b))

## 1.0.0 (2026-07-24)


### Features

* **auth:** config/token store with 0600 perms ([4170ed6](https://github.com/regask/backstage-cli/commit/4170ed64244dd2b040f7df164110a67f8e650a97))
* **auth:** loopback login, logout, whoami ([b97d64c](https://github.com/regask/backstage-cli/commit/b97d64cfca0be4dd690d549e621ad55d26837795))
* check-deploy and check-environment commands ([287fef8](https://github.com/regask/backstage-cli/commit/287fef8c8f1dea288ae9de95546a7faab2ed3334))
* check-secrets via gitops overlay + local az (masked by default) ([7de8684](https://github.com/regask/backstage-cli/commit/7de8684a2b872595bf5bba25f92d7d61d35bbffb))
* **cli:** consent pairing code on login; cherry-pick --tag is a ticket, --branch limited to release/preprod|prod ([f67fca7](https://github.com/regask/backstage-cli/commit/f67fca72e964c284bb949ae5eb7b3323b7f9f1f3))
* **client:** typed HTTP client with bearer + fresh + 401 sentinel ([f7753c0](https://github.com/regask/backstage-cli/commit/f7753c02bcae609ff847e32c694c4396769d0aa4))
* **contracts:** env/secret parsers and backend JSON types ([27ba54f](https://github.com/regask/backstage-cli/commit/27ba54f4da60239208131ae598eef0f0fb56123c))
* find-ticket command ([b3c8dee](https://github.com/regask/backstage-cli/commit/b3c8dee7bd13b2be68911923fe7aa6a1c0c3f895))
* **login:** point at portal /cli-auth handshake + CSRF state verification ([c07bf37](https://github.com/regask/backstage-cli/commit/c07bf3796ca96f3303b5821fffae777d5b010209))
* promote/release/cherry-pick launch scaffolder templates with live log ([236128b](https://github.com/regask/backstage-cli/commit/236128bffbed8e11221cce1fe55df63a95439c5b))
* query-approval + approve/reject with release link and task backlink ([bbcdbb8](https://github.com/regask/backstage-cli/commit/bbcdbb8f9167af95b90b689ec906f17a4b9267f8))
* scaffold cobra CLI with --json and render helper ([6113511](https://github.com/regask/backstage-cli/commit/61135110d32536672eb7af7321f2e313f7b3a100))


### Bug Fixes

* approve prints 'rejected' not 'rejectd' for --reject ([d79f302](https://github.com/regask/backstage-cli/commit/d79f3027ccdb2bca6ccd7ba4e3ac240fbe0f0f0a))
* **auth:** enforce 0700/0600 on pre-existing config dir/file ([99940ae](https://github.com/regask/backstage-cli/commit/99940aeba2c1406f94056f8157df2bdd47200b61))
* **contracts:** anchor remoteRef match to keep parseSecretRefs parity with backend ([32f8668](https://github.com/regask/backstage-cli/commit/32f86685dda782eea8523eb7902402f0a7cc9b91))

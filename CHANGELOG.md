# Changelog

## [1.2.0](https://github.com/regask/backstage-cli/compare/v1.1.0...v1.2.0) (2026-07-25)


### Features

* **approvals:** prefer payload.draftReleaseUrl for the release link when present ([cc3cf05](https://github.com/regask/backstage-cli/commit/cc3cf05c7a988cc62adb6c3f946f4578c6ef25fe))
* **approvals:** show draft release link (from payload) while pending; published resultUrl when done ([b5bbb29](https://github.com/regask/backstage-cli/commit/b5bbb290f85cbd9f42391b02cd84a5ca0650ca8a))
* **tui:** add bubbletea deps, theme, and key bindings ([4383fd8](https://github.com/regask/backstage-cli/commit/4383fd8f3b6ab7bd3819c8f2cf3e9ab74a39b5ef))
* **tui:** approvals count badge in header + services unhealthy-only filter (u) ([9fc260a](https://github.com/regask/backstage-cli/commit/9fc260abacdaa397f3d24d17c924ec53675907f5))
* **tui:** approvals view with detail pane ([d7a4844](https://github.com/regask/backstage-cli/commit/d7a484417e5a7fad0fd3cd16373863f0f56deb47))
* **tui:** bsr ui command launching the Bubble Tea program ([ebf3127](https://github.com/regask/backstage-cli/commit/ebf3127b2f823f93ce8995409c0c49b89947bf0c))
* **tui:** command bar and status header/footer ([92d668b](https://github.com/regask/backstage-cli/commit/92d668bc5a38e205ff0510bb0f6168dc50a0808f))
* **tui:** message types and async data commands ([990d8a8](https://github.com/regask/backstage-cli/commit/990d8a8d4086396c5045ec21fdb0091becf59ce5))
* **tui:** promote/release/approve/reject actions with confirm ([888e283](https://github.com/regask/backstage-cli/commit/888e28327b82efc9a58e3178145d611fdc4a9331))
* **tui:** root app model wiring views, command bar, global keys ([4f1a987](https://github.com/regask/backstage-cli/commit/4f1a9874c2e6c2dc124ca5bd4dfc63782080fb1d))
* **tui:** services (deploy matrix) view with filter ([4587410](https://github.com/regask/backstage-cli/commit/45874101b77d17f1db7a456c37b090eb8287220a))
* **tui:** Tab toggles between services and approvals views ([67b0d77](https://github.com/regask/backstage-cli/commit/67b0d77321e78180baca8168be8c860329670f19))


### Bug Fixes

* **tui:** correct services Selected() row mapping; use keys.Filter; color sync/health ([5422e85](https://github.com/regask/backstage-cli/commit/5422e8511a42c2f7f6beeb3add6145301bbb071c))
* **tui:** guard view input focus, constrain layout to viewport, theme selected row + banners ([86b4055](https://github.com/regask/backstage-cli/commit/86b4055de1550cf63592b7dccba0cb011dc9f8c0))

## [1.1.0](https://github.com/regask/backstage-cli/compare/v1.0.3...v1.1.0) (2026-07-24)


### Features

* **check-secrets:** default vault per env (SECRET_VAULT_BY_ENV); --vault now an override ([fd3f20e](https://github.com/regask/backstage-cli/commit/fd3f20ee292c264e8da04b0adced0c7ef7299cc6))


### Bug Fixes

* **approvals:** unwrap { request } envelope in GetApproval (query-approval showed empty) ([138053e](https://github.com/regask/backstage-cli/commit/138053e602834b985f9c23dcaca3e394ea1c862d))

## [1.0.3](https://github.com/regask/backstage-cli/compare/v1.0.2...v1.0.3) (2026-07-24)


### Bug Fixes

* promote/release resolve bare service names to entity refs (consistent with check-deploy) ([c0a8348](https://github.com/regask/backstage-cli/commit/c0a8348f3b34f2063792ba67470430a102978aae))

## [1.0.2](https://github.com/regask/backstage-cli/compare/v1.0.1...v1.0.2) (2026-07-24)


### Bug Fixes

* resolve bare service name to entity ref (matrix/overlays exact-match ref); add ArgoCD sync/health columns to check-deploy ([844a80f](https://github.com/regask/backstage-cli/commit/844a80f0f6b103d9270456d3087494e40ba4ba59))

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

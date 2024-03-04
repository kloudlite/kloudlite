# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- mongodb standalone controller was missing events generated on `HelmResource` it creates, which caused mongodb standalone services, not getting ready,or hanging while being deleted.

- fixes project managed service namespace getting deleted before its child resources getting properly finalized

## [v1.0.2] - 2024-02-13

### Added

- cluster and nodepool job pods, now have desired observability annotations for log and metrics scraping

## [v1.0.1] - 2024-02-07

### Added

- [operators/routers] router now gets ready, when there are no routes

## [v1.0.0] - 2024-01-26

### Added

- [CI Build Operators](./.github/workflows/build-operators.yml) job step `Build & Push Image` has been updated to not force push docker images for `non-nightly` branches. This way, our production assets will be immutable.
- router controller now adds annotation `nginx.ingress.kubernetes.io/from-to-www-redirect: "true"` to all the ingress resources, by default
- helmchart CR, now has `.spec.releaseName` optional field to have a different release name than `.metadata.name`.

[v1.0.1]: https://github.com/kloudlite/operator/compare/v1.0.0...v1.0.1
[v1.0.0]: https://github.com/kloudlite/operator/compare/master...release-1.0.0

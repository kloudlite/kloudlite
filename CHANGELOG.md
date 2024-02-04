# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v1.0.5] - 2024-01-26

### Added

- [CI Build Operators](./.github/workflows/build-operators.yml) job step `Build & Push Image` has been updated to not force push docker images for `non-nightly` branches. This way, our production assets will be immutable.
- router controller now adds annotation `nginx.ingress.kubernetes.io/from-to-www-redirect: "true"` to all the ingress resources, by default
- helmchart CR, now has `.spec.releaseName` optional field to have a different release name than `.metadata.name`.

[v1.0.5]: https://github.com/kloudlite/operator/compare/master...release-1.0.5

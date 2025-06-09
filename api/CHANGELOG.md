# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v1.0.1] - 2024-02-07

### Added

- [apps/console] fixes managed resources created during environment cloning, `.spec.resourceName` is now generated differently for cloned environment
- [apps/iam] fixes resolution of role `account-member` for actions `read-logs`, and `read-metrics`
- [apps/infra] adds support for PV deletion
- [apps/infra] fixes `getDevice` API. In case of unavailablity of wireguard config, it threw error, which caused [kloudlite/kl] to exit with non-zero code.

## [v1.0.0] - 2024-02-04

### Added 

- [apps/infra] tenant clusters installation of `charts/kloudlite-agent` is now installed and managed by infra API. It is done to ensure that kloudlite can upgrade those releases, as new releases arrive

[v1.0.1]: https://github.com/kloudlite/api/compare/v1.0.0...v1.0.1
[kloudite/kl]: https://github.com/kloudlite/ki

# Release Versioning Policy

This repository uses artifact-scoped version tags.

## Baseline

- Initial baseline for all stable releases: `0.1.0`
- Stable tags use SemVer: `<artifact>-v<MAJOR>.<MINOR>.<PATCH>`
- Nightly builds use pre-release format from workflows:
  - `0.1.0-nightly.<UTC_TIMESTAMP>.<SHORT_SHA>`

## Tag Prefixes

- `kli-v*`
- `kltun-v*`
- `api-server-v*`
- `code-analyzer-v*`
- `k3s-backup-v*`
- `nix-image-v*`
- `workmachine-node-manager-v*`
- `oci-installer-v*`
- `tunnel-server-v*`
- `wm-ingress-controller-v*`
- `workspace-images-v*`
- `web-console-v*`
- `web-dashboard-v*`
- `web-website-v*`

## Examples

- `kli-v0.1.0`
- `kltun-v0.1.1`
- `api-server-v0.2.0`
- `web-website-v0.1.3`

## Notes

- Tag-triggered workflows only process their own prefixed tags.
- Nightly releases for binaries are scheduled in GitHub Actions and marked as prerelease.
- Deployments should prefer immutable `sha-*` image tags or image digests for GitOps promotion.

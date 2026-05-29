# Build, Release, and Deploy Flow

## Repos

- `kloudlite/kloudlite`
  - builds Docker images and binaries
  - creates GitHub releases/prereleases for CLI and desktop artifacts
  - triggers nightly deploys
- `kloudlite/kloudlite-ci`
  - runs the real AKS deploy jobs
  - owns the Helm chart and deploy-time environment secrets
  - single branch: `master`

## Intended Policy

- Feature branches
  - on every push, build only the apps affected by that change
- `development`
  - every night, build everything from `development`
  - publish nightly artifacts where applicable
  - deploy nightly to `development`
- `uat`
  - per-app RC tags should create RC builds/releases
- `master`
  - per-app production tags should create production builds/releases

Per-app tag examples:

- `console/v1.2.0`
- `platform-controller/v2.1.0`
- `kli/v1.5.0`
- `desktop/v0.4.0`

## What Happens Today

### Feature Branch Pushes

Source: `.github/workflows/build-on-push.yml`

- Trigger: pushes to all branches except `master`, `uat`, and `development`
- Tag: `<branch>-<short-sha>`
- Behavior: builds only affected apps

Current mapping:

| Change | Build |
|---|---|
| `api/**` | all Go Docker apps |
| `web/apps/console/**` or `web/packages/**` | `console` |
| `web/apps/dashboard/**` or `web/packages/**` | `dashboard` |
| `web/apps/website/**` or `web/packages/**` | `website` |
| `api/cmd/kli/**` and shared Go files | `kli` |
| `api/cmd/kltun/**` and shared Go files | `kltun` |
| `web/apps/desktop/**` | `desktop` |

### Nightly

Source: `.github/workflows/build-nightly.yml`

| Item | Value |
|---|---|
| Schedule | `27 18 * * *` |
| Docker builds | all Go Docker apps, `console`, `dashboard`, `website` |
| Binary/Desktop builds | `kli`, `kltun`, `desktop` |
| Nightly prereleases | `kli`, `kltun`, `desktop` |
| Deploy target | `development` through `kloudlite-ci` |

Nightly naming:

- tag: `nightly-<YYYYMMDD>-<short-sha>`
- binary version: `0.0.0-nightly.<YYYYMMDD>`

Important:

- GitHub scheduled workflows run from the repo default branch
- nightly is only guaranteed to use `development` if the default branch is `development`, or the workflow is changed to force that branch

### Tag Releases

Source: `.github/workflows/build-release.yml`

| Item | Value |
|---|---|
| Trigger | any tag matching `*/v*` |
| Tag format | `<app>/v<version>` |
| Behavior | builds only the tagged app |

Current release behavior:

| App type | Current behavior |
|---|---|
| `kli` | build + GitHub release |
| `kltun` | build + GitHub release |
| Docker apps | build/publish image only |
| `desktop` | not yet included in tag release flow |

## Deploy Flow

Nightly deploy path:

1. `kloudlite` builds artifacts/images
2. `kloudlite` calls `kloudlite-ci/.github/workflows/deploy.yml`
3. `kloudlite-ci` logs into AKS and runs Helm

Important:

- the real deploy job runs in `kloudlite-ci`, not in `kloudlite`

## Environment Secrets

`kloudlite-ci` recreates `console-secrets` during deploy from GitHub environment secrets.

Required secrets:

- `SUPABASE_URL`
- `SUPABASE_KEY`
- `PII_SUPABASE_URL`
- `PII_SUPABASE_KEY`

Current environment mapping:

| Environment | Database mapping |
|---|---|
| `development` | development console DB + development PII DB |
| `staging` | production console DB + production PII DB |
| `production` | production console DB + production PII DB |

## Gaps

These policy rules are not fully enforced yet:

- nightly is not explicitly pinned to `development`
- RC releases are not explicitly restricted to `uat`
- production releases are not explicitly restricted to `master`
- `desktop` is in nightly release flow, but not yet in tag release flow

## Quick Examples

| Scenario | Example | Result |
|---|---|---|
| Feature branch web change | branch `feat/improve-console-billing`, files in `web/apps/console/**` | `console` build runs with tag like `feat-improve-console-billing-1a2b3c4` |
| Feature branch API change | branch `fix/tunnel-timeout`, files in `api/pkg/**` | all Go Docker apps rebuild |
| Nightly | commit `abcdef123456...`, date `2026-04-08` | tag `nightly-20260408-abcdef1`, version `0.0.0-nightly.20260408`, full build + nightly prereleases + deploy to `development` |
| Intended UAT RC release | branch `uat`, tag `console/v1.4.0-rc.1` | RC build/release only |
| Intended master production release | branch `master`, tag `console/v1.4.0` | production build/release only |

## Fast Mental Model

- feature branches -> build changed apps only
- nightly -> build everything, publish nightly artifacts, deploy to `development`
- tags -> release one app at a time
- deploys -> always executed in `kloudlite-ci`

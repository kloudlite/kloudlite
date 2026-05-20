# build-sys

`build-sys` is the single home for Kloudlite Docker image definitions, local image build tasks, and image documentation.

## Commands

Run from the repository root:

```bash
task build-sys:image:api-server
task build-sys:image:workmachine-manager
task build-sys:image:all
```

Override defaults:

```bash
task build-sys:image:api-server TAG=dev-local REGISTRY=ghcr.io/kloudlite/kloudlite PLATFORM=linux/amd64
```

Defaults:

- `REGISTRY=ghcr.io/kloudlite/kloudlite`
- `TAG=dev-local`
- `PLATFORM=linux/amd64`

## Image catalog

| Image | Dockerfile | Context | Pre-build artifact | Notes |
| --- | --- | --- | --- | --- |
| api-server | `build-sys/images/api-server/Dockerfile` | repo root | `bin/api-server` | API server runtime image. |
| workmachine-manager | `build-sys/images/workmachine-manager/Dockerfile` | repo root | `bin/workmachine-manager` | Per-WorkMachine manager runtime image. |
| nix | `build-sys/images/nix/Dockerfile` | repo root | none | Nix helper image used to seed shared Nix volumes. |
| nix-serve | `build-sys/images/nix-serve/Dockerfile` | repo root | none | Nix binary cache server image. |
| k3s-backup | `build-sys/images/k3s-backup/Dockerfile` | repo root | none | K3s backup support image. |
| workspace-base | `build-sys/images/workspace-base/Dockerfile` | `api/workspace-images/base` | none | Base workspace image. |
| workspace-comprehensive | `build-sys/images/workspace-comprehensive/Dockerfile` | repo root | `bin/kl-linux` | Full workspace image built from `workspace-base`. |
| code-analyzer | `build-sys/images/code-analyzer/Dockerfile` | repo root | `bin/code-analyzer` | Semgrep/code-analysis support image. |
| oci-installer | `build-sys/images/oci-installer/Dockerfile` | repo root | `bin/oci-installer` | OCI installer image. |
| kl-tun-proxy | `build-sys/images/kl-tun-proxy/Dockerfile` | repo root | built inside Dockerfile | Tunnel proxy image. |
| console | `build-sys/images/web-console/Dockerfile.runtime` | `web/apps/console` | app build output | Console runtime image. |
| dashboard | `build-sys/images/web-dashboard/Dockerfile.runtime` | `web/apps/dashboard` | app build output | Dashboard runtime image. |
| website | `build-sys/images/web-website/Dockerfile.runtime` | `web/apps/website` | app build output | Website runtime image. |

## Conventions

- Keep Dockerfiles under `build-sys/images/<image-name>/` only.
- Keep image-specific notes in `build-sys/images/<image-name>/README.md`.
- Prefer repo-root build context when an image copies files from multiple project areas.
- Keep existing image suffixes stable unless deployment manifests are updated in the same change.

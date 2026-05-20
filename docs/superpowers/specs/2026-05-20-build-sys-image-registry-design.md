# build-sys Image Registry Design

## Goal

Create a single, documented home for every Docker image definition in the repository. The new `build-sys/` directory will own image Dockerfiles, image build tasks, and image documentation. Old Dockerfile locations will be removed to avoid confusion.

## Scope

Move all current Dockerfiles into `build-sys/images/<image-name>/`:

- `api-server`
- `workmachine-manager`
- `nix`
- `nix-serve`
- `k3s-backup`
- `workspace-base`
- `workspace-comprehensive`
- `code-analyzer`
- `oci-installer`
- `kl-tun-proxy`
- `web-console`
- `web-dashboard`
- `web-website`

No compatibility Dockerfile stubs will remain at the old paths.

## Directory layout

```txt
build-sys/
  README.md
  Taskfile.yml
  images/
    api-server/
      Dockerfile
      README.md
    workmachine-manager/
      Dockerfile
      README.md
    nix/
      Dockerfile
      README.md
    nix-serve/
      Dockerfile
      README.md
    k3s-backup/
      Dockerfile
      README.md
    workspace-base/
      Dockerfile
      README.md
    workspace-comprehensive/
      Dockerfile
      README.md
    code-analyzer/
      Dockerfile
      README.md
    oci-installer/
      Dockerfile
      README.md
    kl-tun-proxy/
      Dockerfile
      README.md
    web-console/
      Dockerfile
      Dockerfile.runtime
      README.md
    web-dashboard/
      Dockerfile
      Dockerfile.runtime
      README.md
    web-website/
      Dockerfile
      Dockerfile.runtime
      README.md
```

## Build tasks

`build-sys/Taskfile.yml` will expose image-specific tasks and an aggregate task:

```txt
task build-sys:image:api-server
task build-sys:image:workmachine-manager
task build-sys:image:nix
task build-sys:image:nix-serve
task build-sys:image:k3s-backup
task build-sys:image:workspace-base
task build-sys:image:workspace-comprehensive
task build-sys:image:code-analyzer
task build-sys:image:oci-installer
task build-sys:image:kl-tun-proxy
task build-sys:image:console
task build-sys:image:dashboard
task build-sys:image:website
task build-sys:image:all
```

The root `Taskfile.yml` will include `build-sys/Taskfile.yml` under the `build-sys` namespace while keeping the existing `api` include.

## Build contexts

Dockerfiles move, but build contexts stay image-appropriate:

- Go runtime/support images use repo root context when they copy binaries from `bin/` or source paths.
- `workspace-base` keeps its existing workspace-base context.
- `workspace-comprehensive` keeps repo root context because it consumes `bin/kl-linux` and accepts `BASE_IMAGE`.
- Web app images keep their app context unless a Dockerfile requires root-level files.

## CI updates

`.github/workflows/_build-docker.yml` will be updated to reference Dockerfiles under `build-sys/images/...`. Image names, tags, build commands, and contexts remain unchanged unless required by the new Dockerfile locations.

Path filters in push/nightly/release workflows should include `build-sys/**` so image-definition changes trigger the relevant builds.

## Documentation

`build-sys/README.md` will document:

- image catalog
- build command examples
- tagging conventions
- source binary/app for each image
- build context for each image

Each image README will document:

- purpose
- source Dockerfile origin
- build context
- output image suffix
- required pre-build artifacts, if any
- runtime notes

## Verification

Verification will include:

- `task --list` to confirm task registration.
- Focused dry/path checks for moved Dockerfiles.
- At least one representative local Docker build for a lightweight image if Docker is available.
- Go/web tests are not required for a pure file move unless code paths are changed.

## Runtime impact

This change should not alter image contents by itself. Runtime impact is limited to build-system paths and CI references. Deployed environments are unaffected until images are rebuilt and deployed.

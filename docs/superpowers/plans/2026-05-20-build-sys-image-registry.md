# build-sys Image Registry Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move every Docker image definition and image build task into a documented `build-sys/` directory, remove old Dockerfile paths, update CI, and open a PR.

**Architecture:** `build-sys/images/<image-name>/` becomes the single registry for Dockerfiles and image documentation. `build-sys/Taskfile.yml` owns image build commands, while the root `Taskfile.yml` includes it under the `build-sys` namespace. GitHub Docker workflows point at the new paths while preserving existing image names, tags, build commands, and build contexts.

**Tech Stack:** Docker/BuildKit, Taskfile v3, GitHub Actions, Go binaries, Bun/Next.js web images.

---

## File structure

- Create `build-sys/README.md`: central image catalog, build commands, conventions.
- Create `build-sys/Taskfile.yml`: per-image local Docker build tasks and aggregate image build tasks.
- Create `build-sys/images/*/README.md`: per-image purpose, context, source artifact, output image, runtime notes.
- Move existing Dockerfiles into:
  - `build-sys/images/api-server/Dockerfile`
  - `build-sys/images/workmachine-manager/Dockerfile`
  - `build-sys/images/nix/Dockerfile`
  - `build-sys/images/nix-serve/Dockerfile`
  - `build-sys/images/k3s-backup/Dockerfile`
  - `build-sys/images/workspace-base/Dockerfile`
  - `build-sys/images/workspace-comprehensive/Dockerfile`
  - `build-sys/images/code-analyzer/Dockerfile`
  - `build-sys/images/oci-installer/Dockerfile`
  - `build-sys/images/kl-tun-proxy/Dockerfile`
  - `build-sys/images/web-console/Dockerfile`
  - `build-sys/images/web-console/Dockerfile.runtime`
  - `build-sys/images/web-dashboard/Dockerfile`
  - `build-sys/images/web-dashboard/Dockerfile.runtime`
  - `build-sys/images/web-website/Dockerfile`
  - `build-sys/images/web-website/Dockerfile.runtime`
- Modify `Taskfile.yml`: include `build-sys/Taskfile.yml` under `build-sys`.
- Modify `.github/workflows/_build-docker.yml`: replace Dockerfile paths with `build-sys/images/...` paths.
- Modify `.github/workflows/build-on-push.yml`: add `build-sys/**` to relevant change filters.
- Modify `.github/workflows/build-nightly.yml` only if required by path references; nightly calls reusable workflow by app name and should not need path changes.

## Task 1: Move Dockerfiles into build-sys

**Files:**
- Create directories under `build-sys/images/`.
- Move every Dockerfile from old source locations to new image directories.

- [ ] **Step 1: Create image directories**

Run:

```bash
mkdir -p build-sys/images/{api-server,workmachine-manager,nix,nix-serve,k3s-backup,workspace-base,workspace-comprehensive,code-analyzer,oci-installer,kl-tun-proxy,web-console,web-dashboard,web-website}
```

Expected: command exits 0.

- [ ] **Step 2: Move Dockerfiles**

Run:

```bash
mv api/Dockerfile.api-server build-sys/images/api-server/Dockerfile
mv api/Dockerfile.workmachine-manager build-sys/images/workmachine-manager/Dockerfile
mv api/Dockerfile.nix build-sys/images/nix/Dockerfile
mv api/Dockerfile.nix-serve build-sys/images/nix-serve/Dockerfile
mv api/Dockerfile.k3s-backup build-sys/images/k3s-backup/Dockerfile
mv api/workspace-images/base/Dockerfile build-sys/images/workspace-base/Dockerfile
mv api/workspace-images/comprehensive/Dockerfile build-sys/images/workspace-comprehensive/Dockerfile
mv cli/code-analyzer/Dockerfile build-sys/images/code-analyzer/Dockerfile
mv cli/kli/oci-installer/Dockerfile build-sys/images/oci-installer/Dockerfile
mv cli/kl-tun-proxy/Dockerfile build-sys/images/kl-tun-proxy/Dockerfile
mv web/apps/console/Dockerfile build-sys/images/web-console/Dockerfile
mv web/apps/console/Dockerfile.runtime build-sys/images/web-console/Dockerfile.runtime
mv web/apps/dashboard/Dockerfile build-sys/images/web-dashboard/Dockerfile
mv web/apps/dashboard/Dockerfile.runtime build-sys/images/web-dashboard/Dockerfile.runtime
mv web/apps/website/Dockerfile build-sys/images/web-website/Dockerfile
mv web/apps/website/Dockerfile.runtime build-sys/images/web-website/Dockerfile.runtime
```

Expected: command exits 0 and old Dockerfile paths no longer exist.

- [ ] **Step 3: Verify no old Dockerfile files remain**

Run:

```bash
git status --short
```

Expected: old Dockerfile paths are shown as deleted and new `build-sys/images/...` paths are shown as added.

## Task 2: Add build-sys Taskfile

**Files:**
- Create: `build-sys/Taskfile.yml`
- Modify: `Taskfile.yml`

- [ ] **Step 1: Create `build-sys/Taskfile.yml`**

Add tasks using repo-root working directory assumptions:

```yaml
version: 3

vars:
  REGISTRY: '{{.REGISTRY | default "ghcr.io/kloudlite/kloudlite"}}'
  TAG: '{{.TAG | default "dev-local"}}'
  PLATFORM: '{{.PLATFORM | default "linux/amd64"}}'

tasks:
  image:api-server:
    desc: Build api-server image
    deps:
      - task: ../api:build:api-server
    cmds:
      - docker buildx build --platform {{.PLATFORM}} -f build-sys/images/api-server/Dockerfile -t {{.REGISTRY}}/api-server:{{.TAG}} --load .

  image:workmachine-manager:
    desc: Build workmachine-manager image
    deps:
      - task: ../api:build:workmachine-manager
    cmds:
      - docker buildx build --platform {{.PLATFORM}} -f build-sys/images/workmachine-manager/Dockerfile -t {{.REGISTRY}}/workmachine-manager:{{.TAG}} --load .

  image:nix:
    desc: Build nix helper image
    cmds:
      - docker buildx build --platform {{.PLATFORM}} -f build-sys/images/nix/Dockerfile -t {{.REGISTRY}}/nix:{{.TAG}} --load .

  image:nix-serve:
    desc: Build nix-serve image
    cmds:
      - docker buildx build --platform {{.PLATFORM}} -f build-sys/images/nix-serve/Dockerfile -t {{.REGISTRY}}/nix-serve:{{.TAG}} --load .

  image:k3s-backup:
    desc: Build k3s-backup image
    cmds:
      - docker buildx build --platform {{.PLATFORM}} -f build-sys/images/k3s-backup/Dockerfile -t {{.REGISTRY}}/k3s-backup:{{.TAG}} --load .

  image:workspace-base:
    desc: Build workspace-base image
    cmds:
      - docker buildx build --platform {{.PLATFORM}} -f build-sys/images/workspace-base/Dockerfile -t {{.REGISTRY}}/workspace-base:{{.TAG}} --load api/workspace-images/base

  image:workspace-comprehensive:
    desc: Build workspace-comprehensive image
    cmds:
      - CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -trimpath -o bin/kl-linux cli/kl/main.go
      - docker buildx build --platform {{.PLATFORM}} -f build-sys/images/workspace-comprehensive/Dockerfile --build-arg BASE_IMAGE={{.REGISTRY}}/workspace-base:{{.TAG}} --build-arg VERSION={{.TAG}} -t {{.REGISTRY}}/workspace-comprehensive:{{.TAG}} --load .

  image:code-analyzer:
    desc: Build code-analyzer image
    cmds:
      - CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o ./bin/code-analyzer ./cli/code-analyzer
      - docker buildx build --platform {{.PLATFORM}} -f build-sys/images/code-analyzer/Dockerfile -t {{.REGISTRY}}/code-analyzer:{{.TAG}} --load .

  image:oci-installer:
    desc: Build oci-installer image
    cmds:
      - CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o ./bin/oci-installer ./cli/kli/oci-installer
      - docker buildx build --platform {{.PLATFORM}} -f build-sys/images/oci-installer/Dockerfile -t {{.REGISTRY}}/oci-installer:{{.TAG}} --load .

  image:kl-tun-proxy:
    desc: Build kl-tun-proxy image
    cmds:
      - docker buildx build --platform {{.PLATFORM}} -f build-sys/images/kl-tun-proxy/Dockerfile -t {{.REGISTRY}}/kl-tun-proxy:{{.TAG}} --load .

  image:console:
    desc: Build console runtime image
    cmds:
      - docker buildx build --platform {{.PLATFORM}} -f build-sys/images/web-console/Dockerfile.runtime -t {{.REGISTRY}}/console:{{.TAG}} --load web/apps/console

  image:dashboard:
    desc: Build dashboard runtime image
    cmds:
      - docker buildx build --platform {{.PLATFORM}} -f build-sys/images/web-dashboard/Dockerfile.runtime -t {{.REGISTRY}}/dashboard:{{.TAG}} --load web/apps/dashboard

  image:website:
    desc: Build website runtime image
    cmds:
      - docker buildx build --platform {{.PLATFORM}} -f build-sys/images/web-website/Dockerfile.runtime -t {{.REGISTRY}}/website:{{.TAG}} --load web/apps/website

  image:all:
    desc: Build all Docker images except dependent workspace-comprehensive ordering is explicit
    cmds:
      - task: image:api-server
      - task: image:workmachine-manager
      - task: image:nix
      - task: image:nix-serve
      - task: image:k3s-backup
      - task: image:workspace-base
      - task: image:workspace-comprehensive
      - task: image:code-analyzer
      - task: image:oci-installer
      - task: image:kl-tun-proxy
      - task: image:console
      - task: image:dashboard
      - task: image:website
```

- [ ] **Step 2: Include build-sys in root Taskfile**

Modify `Taskfile.yml` includes:

```yaml
includes:
  api:
    taskfile: ./api/Taskfile.yml
    dir: .
  build-sys:
    taskfile: ./build-sys/Taskfile.yml
    dir: .
```

- [ ] **Step 3: Verify task registration**

Run:

```bash
task --list
```

Expected: output includes `build-sys:image:api-server`, `build-sys:image:workmachine-manager`, and `build-sys:image:all`.

## Task 3: Update GitHub Docker workflow paths

**Files:**
- Modify: `.github/workflows/_build-docker.yml`
- Modify: `.github/workflows/build-on-push.yml`

- [ ] **Step 1: Update Dockerfile path map in `_build-docker.yml`**

Replace old Dockerfile assignments with:

```yaml
DOCKERFILE="./build-sys/images/oci-installer/Dockerfile"
DOCKERFILE="./build-sys/images/code-analyzer/Dockerfile"
DOCKERFILE="./build-sys/images/k3s-backup/Dockerfile"
DOCKERFILE="./build-sys/images/workspace-base/Dockerfile"
DOCKERFILE="./build-sys/images/workspace-comprehensive/Dockerfile"
DOCKERFILE="./build-sys/images/web-console/Dockerfile.runtime"
DOCKERFILE="./build-sys/images/web-dashboard/Dockerfile.runtime"
DOCKERFILE="./build-sys/images/web-website/Dockerfile.runtime"
```

Keep existing `DOCKER_CONTEXT` values unless a build fails from missing files.

- [ ] **Step 2: Add build-sys to push change filters**

In `.github/workflows/build-on-push.yml`, add `build-sys/**` to filters that should rebuild affected images. Since image definitions can affect every image, add it to `api`, `console`, `dashboard`, `website`, `kli`, and `kltun` filters.

- [ ] **Step 3: Check old path references**

Run:

```bash
rg 'api/Dockerfile|web/apps/.*/Dockerfile|cli/.*/Dockerfile|api/workspace-images/.*/Dockerfile' .github build-sys Taskfile.yml
```

Expected: no stale Dockerfile references except documentation that explicitly describes old origins.

## Task 4: Document build-sys and each image

**Files:**
- Create: `build-sys/README.md`
- Create: `build-sys/images/*/README.md`

- [ ] **Step 1: Write top-level docs**

`build-sys/README.md` must include:

```markdown
# build-sys

`build-sys` is the single home for Kloudlite Docker image definitions and image build tasks.

## Commands

- `task build-sys:image:api-server`
- `task build-sys:image:workmachine-manager`
- `task build-sys:image:all`

Override defaults with:

```bash
task build-sys:image:api-server TAG=dev-local REGISTRY=ghcr.io/kloudlite/kloudlite
```

## Image catalog

| Image | Dockerfile | Context | Notes |
| --- | --- | --- | --- |
| api-server | `build-sys/images/api-server/Dockerfile` | repo root | Copies `bin/api-server`. |
| workmachine-manager | `build-sys/images/workmachine-manager/Dockerfile` | repo root | Copies `bin/workmachine-manager`. |
| nix | `build-sys/images/nix/Dockerfile` | repo root | Nix helper image. |
| nix-serve | `build-sys/images/nix-serve/Dockerfile` | repo root | Nix binary cache server. |
| k3s-backup | `build-sys/images/k3s-backup/Dockerfile` | repo root | K3s backup support image. |
| workspace-base | `build-sys/images/workspace-base/Dockerfile` | `api/workspace-images/base` | Base workspace image. |
| workspace-comprehensive | `build-sys/images/workspace-comprehensive/Dockerfile` | repo root | Full workspace image; needs `bin/kl-linux`. |
| code-analyzer | `build-sys/images/code-analyzer/Dockerfile` | repo root | Code analyzer support image. |
| oci-installer | `build-sys/images/oci-installer/Dockerfile` | repo root | OCI installer image. |
| kl-tun-proxy | `build-sys/images/kl-tun-proxy/Dockerfile` | repo root | Tunnel proxy image. |
| console | `build-sys/images/web-console/Dockerfile.runtime` | `web/apps/console` | Console runtime image. |
| dashboard | `build-sys/images/web-dashboard/Dockerfile.runtime` | `web/apps/dashboard` | Dashboard runtime image. |
| website | `build-sys/images/web-website/Dockerfile.runtime` | `web/apps/website` | Website runtime image. |
```

- [ ] **Step 2: Write per-image READMEs**

Each README must include the exact image name, Dockerfile path, context, required pre-build artifact, and runtime note. Use the top-level image catalog values.

## Task 5: Verify and commit

**Files:**
- All changed files from Tasks 1-4.

- [ ] **Step 1: Verify Dockerfiles are only under build-sys**

Run:

```bash
find . -name 'Dockerfile*' -print | sort
```

Expected: all project Dockerfiles are under `./build-sys/images/...`.

- [ ] **Step 2: Verify task list**

Run:

```bash
task --list
```

Expected: command succeeds and lists `build-sys:image:*` tasks.

- [ ] **Step 3: Verify workflow references**

Run:

```bash
rg 'Dockerfile' .github build-sys Taskfile.yml
```

Expected: Dockerfile references point to `build-sys/images/...` or are documentation/catalog references.

- [ ] **Step 4: Inspect git diff**

Run:

```bash
git diff --stat && git diff -- build-sys Taskfile.yml .github/workflows/_build-docker.yml .github/workflows/build-on-push.yml
```

Expected: diff shows only the planned build-system move/docs plus spec/plan files.

- [ ] **Step 5: Commit intended files**

Run:

```bash
git add build-sys Taskfile.yml .github/workflows/_build-docker.yml .github/workflows/build-on-push.yml docs/superpowers/specs/2026-05-20-build-sys-image-registry-design.md docs/superpowers/plans/2026-05-20-build-sys-image-registry.md
git add -u api cli web
git commit -m "chore: centralize docker image builds"
```

Expected: commit succeeds. Do not stage unrelated `.antigravitycli/`.

## Task 6: Open PR

**Files:**
- No file changes.

- [ ] **Step 1: Inspect final state**

Run:

```bash
git status --short --branch
git log --oneline -10
```

Expected: branch contains the new build-system commit and no unintended staged files.

- [ ] **Step 2: Push branch and open PR**

Run:

```bash
git push -u origin engineering/configure-azure-workmachine-manifests
gh pr create --base development --head engineering/configure-azure-workmachine-manifests --title "Centralize Docker image builds under build-sys" --body-file /tmp/build-sys-pr.md
```

Expected: GitHub returns a PR URL.

---

## Self-review

- Spec coverage: all Dockerfiles are moved, old paths removed, taskfile added, CI paths updated, docs added, verification and PR steps included.
- Placeholder scan: no placeholder steps remain; each task contains concrete files and commands.
- Type/path consistency: image names and Dockerfile paths match the approved spec.

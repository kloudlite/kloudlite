# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What is Kloudlite?

Cloud development environments with live service connectivity. Developers get per-developer Kubernetes namespaces with workspaces (containers), environments (service namespaces), service interception, and WireGuard tunnels. The platform runs on Kubernetes using controller-runtime CRDs.

## Monorepo Layout

- **`api/`** — Go 1.24 backend. Module: `github.com/kloudlite/kloudlite/api`
- **`web/`** — Bun 1.1 + Turbo monorepo. Next.js 16, React 19, Tailwind 4, TypeScript 5
- **`e2e-tests/`** — Moved to [kloudlite/kloudlite-ci](https://github.com/kloudlite/kloudlite-ci)
- **`supabase/`** — Deno edge functions
- **`manifests/`** — Generated CRDs (do not hand-edit, run `task api:manifests`)

## Essential Commands

### Go (run from `api/`)
```bash
task api:build:server              # Build API server (from repo root)
task api:build:tunnel-server       # Build tunnel server (from repo root)
task api:manifests                 # Regenerate CRDs + deepcopy
cd api && go test ./... -v         # All tests
cd api && go test ./internal/controllers/workspace/ -run TestReconcile -v  # Single test
cd api && go test -v ./pkg/utils/...  # Single package
cd api && golangci-lint run        # Lint (reflect package is BANNED)
gofmt -w <file>                    # Format (enforced by pre-commit hook)
```

### Web (run from `web/`)
```bash
bun install                        # Install deps
bun run dev:console                # Dev single app (also: dev:dashboard, dev:website)
bun run build:console              # Build single app
bun run lint                       # Lint all
bun run format                     # Prettier format all
bunx vitest run src/utils.test.ts  # Run single test file (from package dir)
bunx vitest run -t "test name"     # Run single test by name
```

### E2E Tests

Moved to [kloudlite/kloudlite-ci](https://github.com/kloudlite/kloudlite-ci). Run from that repo's `e2e-tests/` directory.

## Architecture

### Go Backend (`api/`)

**Entrypoints** (`api/cmd/`):
- `server` — Control plane API server (main backend)
- `tunnel-server` — WireGuard tunnel relay for developer connections
- `kl` — CLI that runs inside workspaces (package management, environment switching)
- `kli` — OCI installer CLI
- `kltun` — WireGuard tunnel client (runs on developer's local machine)
- `workmachine-node-manager` — Host-level daemon managing Nix packages on work machine nodes
- `wm-ingress-controller` — Custom ingress controller for work machine routing
- `code-analyzer` — Static analysis service

**Kubernetes Controllers** (`api/internal/controllers/`):
- `workspace/` — Workspace pod lifecycle (create, start, stop, delete)
- `environment/` — Environment namespace management
- `workmachine/` — Work machine node provisioning and health
- `wmingress/` — Ingress routing to work machines
- `composition/` — Composite resource management
- `packages/` — Nix package resolution
- `snapshot/` — Environment snapshots
- `user/` — User resource management

Controllers follow the standard controller-runtime reconciler pattern: each has a `Reconciler` struct, `Reconcile()` method, finalizers for cleanup, owner references for cascading deletes, and label-based filtering.

**Shared Code**:
- `api/internal/` — Internal services, domain logic
- `api/pkg/` — Reusable utilities (exported)

### Web Frontend (`web/`)

**Apps** (`web/apps/`):
- `console` — Main product UI (workspace management, environment config, service intercepts)
- `dashboard` — Account management, billing, team settings
- `website` — Marketing site (kloudlite.io)

**Shared Packages** (`web/packages/`):
- `ui` — shadcn/ui components (import as `@kloudlite/ui`)
- `lib` — Shared utilities, K8s helpers (import as `@kloudlite/lib`)
- `types` — Shared TypeScript types (import as `@kloudlite/types`)

Apps use `@/` alias for local imports (resolves to `src/`). Server actions return `{ success, data?, error? }`.

### CI/CD

Builds happen in this repo. Deployments and e2e tests live in [kloudlite/kloudlite-ci](https://github.com/kloudlite/kloudlite-ci).

#### Branch Strategy

| Branch | Trigger | What builds | Tag format | Deploy |
|--------|---------|-------------|------------|--------|
| **Feature** (not master/uat/development) | Push | Changed apps only (auto-detected) | `<branch>-<sha7>` | None |
| **Development** | Nightly cron (~midnight IST) | All apps | `nightly-YYYYMMDD-<sha7>` + `:nightly` | Auto-deploy to development |
| **UAT** | Manual dispatch | Selected apps (checkboxes) | User-provided version | Manual (from kloudlite-ci) |
| **Master** | `v*` tag push | All apps | Tag as-is (e.g. `v1.2.0`) | Manual (from kloudlite-ci) |

#### Workflow Files (`.github/workflows/`)

**Caller workflows:**
- `build-on-push.yml` — Feature branch builds. Uses `dorny/paths-filter` to detect changes: `api/**` → all Go Docker images, `web/apps/<app>/**` or `web/packages/**` → that web app, `api/cmd/kli/**` or `api/cmd/kltun/**` → CLI binaries
- `build-nightly.yml` — Cron `27 18 * * *` (18:27 UTC). Builds all Docker images + CLI binaries, creates nightly GitHub Releases, then auto-deploys to development via cross-repo reusable workflow (`kloudlite-ci/deploy.yml`)
- `build-release.yml` — `v*` tag trigger (master) + manual dispatch (UAT). Builds all/selected Docker images + CLI binaries, creates GitHub Releases on tag push

**Reusable workflows (prefixed `_`):**
- `_build-docker.yml` — Build single Docker image: Go compile or Bun build → Docker build+push to `ghcr.io/kloudlite/kloudlite/<app>:<tag>`
- `_build-binary.yml` — Cross-compile Go binary (linux/darwin/windows × amd64/arm64) with checksums
- `_release-binary.yml` — Download build artifacts → create GitHub Release with install instructions

#### Docker Images

All pushed to `ghcr.io/kloudlite/kloudlite/<app>:<tag>`:
- **Go:** platform-controller, tunnel-server, workmachine-node-manager, wm-ingress-controller, oci-installer, code-analyzer, k3s-backup, workspace-base, workspace-comprehensive
- **Web:** console, dashboard, website

#### CLI Binaries

`kli` and `kltun` — cross-compiled for 6 platforms, published as GitHub Releases.

#### Deploy (kloudlite-ci)

Deployments are managed from [kloudlite/kloudlite-ci](https://github.com/kloudlite/kloudlite-ci):
- `deploy.yml` — Manual deploy + callable as reusable workflow (cross-repo)
- `deploy-on-build.yml` — Auto-deploy via `repository_dispatch`
- `rollback.yml` — Helm rollback to previous revision
- Helm chart at `helm/kloudlite-apps/` with per-environment values

Nightly auto-deploy uses cross-repo reusable workflow (no PAT needed — kloudlite-ci is public). Requires `AZURE_CREDENTIALS` secret in this repo.

### Infrastructure

AKS cluster `kloudlite` in `rg-kloudlite` (centralindia). NGINX ingress controller at cluster level. Environments map to Kubernetes namespaces (production, staging, development).

## Code Style

### Go

**Imports** — group with blank-line separators: stdlib, external, internal. Alias K8s packages:
```go
import (
    "context"

    ctrl "sigs.k8s.io/controller-runtime"
    apierrors "k8s.io/apimachinery/pkg/api/errors"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

    v1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
)
```

**Error handling** — wrap with context: `fmt.Errorf("failed to parse config: %w", err)`. Use `log.Fatalf` only at startup.

**Patterns** — struct-based services with constructors (`NewWebhookInstaller(...)`), env config via struct tags, `//go:embed` for bundling YAML/static assets.

**Naming** — snake_case files (`environment_controller.go`), `_test.go` suffix co-located.

### TypeScript/React

**Formatting** — no semicolons, single quotes, 2-space indent, trailing commas, max 100 chars. Prettier with `prettier-plugin-tailwindcss` (recognizes `cn` and `clsx`).

**Imports** — `@/` for local, `@kloudlite/{ui,lib,types}` for shared packages, `import type` for type-only, icons from `lucide-react`.

**Components** — `'use client'`/`'use server'` directives, `cva` for variants, `React.forwardRef`, `cn()` for className merging. Forms: `react-hook-form` + `zod`. Toasts: `sonner`.

**Types** — `interface` for object shapes, `type` for unions/intersections. Export Zod schemas with inferred types: `export type Foo = z.infer<typeof fooSchema>`.

**Testing** — Vitest with `happy-dom`, globals enabled. Test files: `*.test.ts` / `*.spec.ts` co-located.

**Naming** — kebab-case files (`work-machine.service.ts`), PascalCase component exports, kebab-case directories.

## Commit Conventions

Conventional Commits: `feat:`, `fix:`, `refactor:`, `docs:`, `chore:`, `perf:`, `ci:`, `style:`. Optional scope: `fix(ui):`, `feat(api):`. Imperative mood, concise summary.

## Key Constraints

- **No `reflect` package** in Go code — denied by depguard linter
- **No semicolons** in TypeScript — Prettier enforced
- **`any` not `interface{}`** in Go — revive linter rule
- **Context must be first parameter** in Go functions
- **Slog messages must be lowercased** with snake_case keys, KV-only (no mixed args)
- **Pre-commit hook** (`.githooks/pre-commit.d/01-go-fmt.sh`) runs `gofmt` on staged Go files
- All Go builds use `CGO_ENABLED=0` and `-ldflags='-s -w'` for static binaries
- Unused vars prefixed with `_` are allowed in TypeScript
- DNS-1123 label validation for Kubernetes resource names in frontend

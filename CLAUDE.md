# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What is Kloudlite?

Cloud development environments with live service connectivity. Developers get per-developer Kubernetes namespaces with workspaces (containers), environments (service namespaces), service interception, and WireGuard tunnels. The platform runs on Kubernetes using controller-runtime CRDs.

## Monorepo Layout

- **`api/`** — Go 1.24 backend. Module: `github.com/kloudlite/kloudlite/api`
- **`web/`** — Bun 1.1 + Turbo monorepo. Next.js 16, React 19, Tailwind 4, TypeScript 5
- **`e2e-tests/`** — Playwright tests
- **`supabase/`** — Deno edge functions
- **`manifests/`** — Generated CRDs (do not hand-edit, run `task api:manifests`)

See `AGENTS.md` for detailed code style, linting rules, and naming conventions.

## Essential Commands

### Go (run from `api/`)
```bash
task api:build:server              # Build API server (from repo root)
task api:manifests                 # Regenerate CRDs + deepcopy
cd api && go test ./... -v         # All tests
cd api && go test ./internal/controllers/workspace/ -run TestReconcile -v  # Single test
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
```

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

Four GitHub Actions workflows:
- `build.yml` — Workflow dispatch: builds selected Docker images (matrix of Go binaries + web apps) → pushes to `ghcr.io/kloudlite/kloudlite/<app>:<tag>`
- `release.yml` — Workflow dispatch: cross-compiles `kli`/`kltun` binaries (linux/darwin/windows × amd64/arm64) → creates GitHub Releases
- `deploy.yml` — Deploys to Azure Container Apps (legacy, per-environment)
- `deploy-aks.yml` — Deploys to AKS via Helm chart (`.github/k8s/kloudlite-apps/`). Uses `--reuse-values` on upgrades so deploying one app doesn't change other apps' image tags

### Infrastructure

AKS cluster `kloudlite` in `rg-kloudlite` (centralindia). NGINX ingress controller at cluster level. Environments map to Kubernetes namespaces (production, staging, development).

Helm chart at `.github/k8s/kloudlite-apps/` with per-environment values files (`values-production.yaml`, etc.).

## Key Constraints

- **No `reflect` package** in Go code — denied by depguard linter
- **No semicolons** in TypeScript — Prettier enforced
- **`any` not `interface{}`** in Go — revive linter rule
- **Context must be first parameter** in Go functions
- **Slog messages must be lowercased** with snake_case keys, KV-only (no mixed args)
- **Pre-commit hook** runs `gofmt` on staged Go files
- All Go builds use `CGO_ENABLED=0` for static binaries

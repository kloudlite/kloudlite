# Project knowledge

## What is this?
Kloudlite — cloud development environments with live service connectivity. Monorepo with a Go API/controller backend and a Next.js web frontend (three apps: console, dashboard, website).

## Project structure
- `api/` — Go backend (API server, K8s controllers, CLI, tunnel server, node manager)
  - `api/cmd/server/` — control plane API server
  - `api/cmd/kl/` — CLI (runs inside workspace)
  - `api/internal/controllers/` — K8s controllers (workspace, environment, serviceintercept)
  - `api/manifests/` — CRDs and RBAC
- `web/` — Bun monorepo (Turborepo) with three Next.js apps + shared packages
  - `web/apps/console/` — console app (port 3002) — billing, installations management
  - `web/apps/dashboard/` — dashboard app (port 3001) — workspace/environment UI
  - `web/apps/website/` — marketing website (port 3000)
  - `web/packages/ui/` — shared UI components (`@kloudlite/ui`)
  - `web/packages/lib/` — shared utilities (`@kloudlite/lib`)
  - `web/packages/types/` — shared types (`@kloudlite/types`)
- `supabase/` — Supabase edge functions (e.g. billing-cron)
- `e2e-tests/` — Playwright end-to-end tests
- `manifests/` — top-level CRD manifests
- `devenv/` — local K3s development setup

## Tech stack
- **Backend:** Go 1.24, controller-runtime, K8s CRDs, WireGuard, Nix
- **Frontend:** Next.js 16, React 19, TypeScript 5, Tailwind CSS 4, Supabase, Razorpay
- **Package manager:** Bun (web monorepo uses `bun` with `workspace:*` protocol)
- **Build orchestration:** Turborepo (`turbo.json` in `web/`)
- **Task runner:** [Task](https://taskfile.dev) (`Taskfile.yml` at root and `api/`)
- **Auth:** next-auth 5 beta, Supabase, jose (JWT)
- **UI:** lucide-react icons, sonner toasts, motion animations, shadcn-style components (`components.json`)

## Commands

### Web (run from `web/`)
```bash
bun install                    # install all web dependencies
bun run dev                    # dev all apps (turbo)
bun run dev:console            # dev console only (port 3002)
bun run dev:dashboard          # dev dashboard only (port 3001)
bun run dev:website            # dev website only (port 3000)
bun run build                  # build all apps
bun run build:console          # build console only
bun run lint                   # lint all apps (currently skipped — Next.js 16 + ESLint 9 compat issue)
bun run format                 # prettier format
```

### API (run from project root or `api/`)
```bash
task api:build:server          # build API server binary
task api:manifests             # generate CRDs and deepcopy
task api:consolidate-crds      # consolidate CRD manifests for CLI
```

### Full dev setup (see SETUP.md)
```bash
cd devenv && task web:install  # install frontend deps
docker-compose up k3s pre-app  # start K3s + pre-setup
task api:dev                   # start backend API (port 8080)
docker-compose up post-app     # post-setup (TLS, users, webhooks)
task web:dev                   # start frontend (port 3000)
```

## Conventions
- **Indentation:** 2 spaces (TS/JS/JSON/YAML/proto), tabs (Go, Makefile)
- **Quotes:** single quotes in TS/JS
- **Line endings:** LF, final newline required
- **Formatting:** Prettier for TS/JS/MD; `gofmt` for Go (pre-commit hook in `.githooks/`)
- **Shared UI:** use `@kloudlite/ui` components; follow shadcn patterns (`components.json`)
- **Shared utils:** use `@kloudlite/lib` (includes `cn()` classname merge utility)
- **Icons:** lucide-react
- **State:** React Query (`@tanstack/react-query`) for server state
- **Forms:** react-hook-form + zod validation
- **Styling:** Tailwind CSS 4 with `tw-animate-css`; use `cn()` from `@kloudlite/lib` for conditional classes
- **Server actions:** Next.js server actions in `src/app/actions/` (console app)
- **Storage layer:** `src/lib/console/storage/` for Supabase DB operations (console app)

## Gotchas
- ESLint is currently disabled in all three web apps due to Next.js 16 + ESLint 9 compatibility issues
- Web package manager is **Bun** (not npm/pnpm) — `bun.lock` exists, `packageManager` field set to `bun@1.1.40`
- Workspace protocol: packages reference each other via `workspace:*`
- React pinned to 19.1.0 via resolutions/overrides
- next-auth is on 5.0.0-beta.30
- Console app uses Razorpay for payments (server-side `razorpay` package)
- Go module path: `github.com/kloudlite/kloudlite/api`

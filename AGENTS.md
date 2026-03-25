# Repository Guidelines

## Project Structure
This monorepo is split by runtime:
- `api/` — Go 1.24 backend (module: `github.com/kloudlite/kloudlite/api`). Entrypoints in `api/cmd/{server,tunnel-server,kli,kl,kltun,...}`, shared logic in `api/internal/` and `api/pkg/`, Kubernetes controllers in `api/internal/controllers/`.
- `web/` — Bun 1.1 + Turbo monorepo. Next.js 16 apps in `web/apps/{website,dashboard,console}`, shared packages in `web/packages/{lib,ui,types}`. React 19, Tailwind 4, TypeScript 5.
- `e2e-tests/` — Playwright end-to-end tests for dashboard and console.
- `supabase/` — Deno edge functions (billing cron, etc.).
- `manifests/` — Generated CRDs and platform manifests.

## Build, Test, and Development Commands

### Go (api/)
- **Build server:** `task api:build:server` (from repo root)
- **Build tunnel server:** `task api:build:tunnel-server` (from repo root)
- **Regenerate CRDs/deepcopy:** `task api:manifests`
- **Run all Go tests:** `cd api && go test ./...`
- **Run a single Go test:** `cd api && go test ./internal/services/ -run TestFunctionName -v`
- **Run tests in one package:** `cd api && go test -v ./pkg/utils/...`
- **Lint Go code:** `cd api && golangci-lint run`
- **Format Go code:** `gofmt -w <file>`

### Web (web/)
- **Install deps:** `cd web && bun install`
- **Dev all apps:** `cd web && bun run dev`
- **Dev single app:** `cd web && bun run dev:console` (or `dev:dashboard`, `dev:website`)
- **Build all:** `cd web && bun run build`
- **Build single app:** `cd web && bun run build:console` (or `build:dashboard`, `build:website`)
- **Lint all:** `cd web && bun run lint`
- **Format:** `cd web && bun run format`
- **Run Vitest (package):** `cd web/packages/lib && bun run test` (or `test:run`, `test:coverage`)
- **Run single Vitest test file:** `cd web/packages/lib && bunx vitest run src/utils.test.ts`
- **Run single test by name:** `cd web/apps/console && bunx vitest run -t "test name pattern"`

### E2E Tests
- **All:** `cd e2e-tests && bun run test`
- **Dashboard only:** `cd e2e-tests && bun run test:dashboard`
- **Console only:** `cd e2e-tests && bun run test:console`
- **Provider tests (long-running):** `bun run test:aws`, `test:gcp`, `test:azure`

## Go Code Style

### Imports
Group imports with blank-line separators: stdlib, external packages, internal packages.
Use aliases for Kubernetes packages:
```go
import (
    "context"
    "fmt"

    "go.uber.org/zap"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"

    apierrors "k8s.io/apimachinery/pkg/api/errors"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

    v1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
)
```

### Error Handling
- Wrap errors with context: `fmt.Errorf("failed to parse config: %w", err)`
- Use `log.Fatalf` only at startup; return errors from functions.
- Structured logging with zap: `logger.Error("reconcile failed", zap.String("name", obj.Name), zap.Error(err))`

### Patterns
- Struct-based services with constructor functions: `NewWebhookInstaller(client, logger, caBundle)`.
- Environment config via struct tags (`envconfig` or `env` tags).
- Controller-runtime reconciler pattern for K8s controllers (finalizers, owner refs, labels).
- `//go:embed` for bundling YAML/static assets.
- Builds use `CGO_ENABLED=0` and `-ldflags='-s -w'` for static binaries.

### Linting (api/.golangci.yml)
Enabled linters: `revive`, `misspell`, `nilerr`, `nilnil`, `sloglint`, `depguard`, `iface`, `unparam`.
- The `reflect` package is **denied** via depguard — do not import it.
- Slog style: snake_case keys, lowercased static messages, context required, KV-only (no mixed args).
- Revive rules enforce: early-return, indent-error-flow, no duplicated imports, context-as-first-argument, use `any` not `interface{}`.
- Formatting enforced by pre-commit hook (`.githooks/pre-commit.d/01-go-fmt.sh`).

## TypeScript/React Code Style

### Formatting & Linting
- **No semicolons.** Single quotes. 2-space indentation. LF line endings. Trailing commas everywhere.
- Max line width: 100 characters.
- Prettier with `prettier-plugin-tailwindcss` for class sorting (recognizes `cn` and `clsx` functions).
- ESLint: `eslint-config-next/core-web-vitals` + `typescript`. Unused vars prefixed with `_` are allowed.

### Imports
- Use `@/` path alias for project-local imports (resolves to `src/`).
- Use `@kloudlite/ui`, `@kloudlite/lib`, `@kloudlite/types` for shared packages.
- Use `import type { ... }` for type-only imports.
- Icons from `lucide-react`.

### Component Patterns
- `'use client'` directive for client components, `'use server'` for server actions.
- shadcn/ui component style: `cva` for variants, `React.forwardRef` for ref forwarding, `cn()` for className merging.
- Forms: `react-hook-form` + `zod` for validation.
- Toasts: `sonner`.
- Server actions return `{ success: boolean, data?: T, error?: string }`.
- Error handling in actions: `try/catch` with `err instanceof Error ? err.message : 'Unknown error'`.

### Types
- Use `interface` for object shapes, `type` for unions/intersections.
- Export Zod schemas and inferred types together: `export type Foo = z.infer<typeof fooSchema>`.
- DNS-1123 label validation for Kubernetes resource names.

### Testing
- Vitest with `happy-dom` environment, globals enabled.
- Coverage via v8 provider (`text`, `json`, `html` reporters).
- Test files: `*.test.ts` / `*.spec.ts` co-located with source.

## Naming Conventions
- **Go files:** snake_case (`environment_controller.go`, `webhook_installer.go`).
- **Go tests:** `_test.go` suffix, co-located with source.
- **TS/JS files:** kebab-case by feature (`work-machine.service.ts`, `create-environment.tsx`).
- **React components:** PascalCase exports, kebab-case filenames.
- **Directories:** lowercase, kebab-case.

## Commit & PR Guidelines
- Conventional Commits: `feat:`, `fix:`, `refactor:`, `docs:`, `chore:`, `perf:`, `ci:`, `style:`.
- Optional scope: `fix(ui):`, `chore(web):`, `feat(api):`.
- Imperative mood, concise summary. Keep commits focused by area.
- PRs: include purpose, impacted modules, test evidence, linked issues, UI screenshots for visual changes.

## Editor Configuration (.editorconfig)
- UTF-8 charset, LF endings, final newline, trim trailing whitespace.
- Go: tabs. TS/JS/JSON/YAML/MD: 2 spaces. Makefiles: tabs.
- Markdown: no trailing whitespace trimming, max 120 chars.
- TS/JS: max 100 chars per line, single quotes.

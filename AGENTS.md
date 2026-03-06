# Repository Guidelines

## Project Structure & Module Organization
This monorepo is split by runtime:
- `api/`: Go backend (API server, CLI tools, controllers, tunnel components). Main entrypoints live in `api/cmd/*`, shared logic in `api/internal/*` and `api/pkg/*`, and Kubernetes manifests in `api/manifests/`.
- `web/`: Bun + Turbo monorepo for Next.js apps and shared packages. Apps are in `web/apps/{website,dashboard,console}` and shared libraries in `web/packages/{lib,ui,types}`.
- `e2e-tests/`: Playwright end-to-end tests for dashboard and console.
- `manifests/`: generated CRDs and platform manifests used across deployments.

## Build, Test, and Development Commands
- `task api:build:server` (repo root): build API server binary.
- `task api:manifests`: regenerate CRDs/deepcopy artifacts.
- `cd api && go test ./...`: run Go unit/integration tests.
- `cd web && bun install && bun run dev`: install and run all web apps via Turbo.
- `cd web && bun run dev:console` (or `dev:dashboard`, `dev:website`): run one app.
- `cd web && bun run lint && bun run build`: lint and production build web workspace.
- `cd e2e-tests && bun run test` or `bun run test:dashboard`: run Playwright suites.

## Coding Style & Naming Conventions
- Follow `.editorconfig`: 2 spaces for TS/JS/MD, tabs for Go, LF endings.
- Always run `gofmt` on Go changes (`.githooks/pre-commit.d/01-go-fmt.sh` enforces this for staged files).
- TS/JS follows ESLint + Prettier in `web/`; use single quotes and keep modules/components focused.
- Use descriptive file names by feature area (example: `work-machine.service.ts`, `parser_test.go`).

## Testing Guidelines
- Go: keep tests next to code with `_test.go`; run `go test ./...` before PRs.
- Web packages/apps use Vitest (`test`, `test:coverage` where available).
- E2E uses Playwright; provider tests (`test:aws|gcp|azure`) are long-running and credential-dependent.
- No global coverage threshold is enforced; new code should include meaningful tests for changed paths.

## Commit & Pull Request Guidelines
- Follow Conventional Commit style seen in history: `feat: ...`, `fix: ...`, `refactor: ...`, `docs: ...`, optional scope (e.g., `style(web): ...`).
- Keep commits focused by area (`api`, `web`, `e2e-tests`) and write imperative summaries.
- PRs should include: purpose, impacted modules, test evidence (commands + results), linked issue(s), and UI screenshots for visual changes.

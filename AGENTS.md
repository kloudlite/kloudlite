# Kloudlite Agent Notes

## Layout That Matters
- The repository root is the Go module (`github.com/kloudlite/kloudlite`). Go code is split across `cli/`, `api/`, `controllers/`, and `pkg/`.
- `web/` is a Bun 1.1 + Turbo workspace. Apps: `console`, `dashboard`, `website`, `desktop`. Shared packages: `@kloudlite/{lib,ui,types}`.
- `manifests/` is generated from Go CRD types. Do not hand-edit; regenerate with `task api:manifests`.
- `e2e-tests/` is not in this repo anymore. `CLAUDE.md` says end-to-end coverage moved to `kloudlite/kloudlite-ci`.

## Commands Agents Usually Guess Wrong
- Root `Taskfile.yml` includes Go tasks under `api`. For web work, run commands from `web/`, not via root `task`.
- Go build commands are run from repo root: `task api:build:kloudlite`, `task api:build:server`, `task api:build:tunnel-server`, `task api:manifests`. Server/tunnel build tasks are aliases for the unified `kloudlite` binary.
- Focused Go verification: `go test ./path/to/pkg -run TestName -v` from the repository root.
- Web workspace commands live in `web/package.json`: `bun run dev`, `dev:console`, `dev:dashboard`, `dev:website`, `dev:desktop`, `build:*`, `lint`, `format`.
- App dev ports are not uniform: `website` 3000, `dashboard` 3001, `console` 3002.
- Shared package lint scripts are placeholders (`echo 'No lint configured...'`). Do not treat `turbo run lint` as meaningful validation for `web/packages/{lib,ui,types}` changes.
- Focused web tests are package-local. Examples: `cd web/packages/lib && bun run test:run`, `cd web/packages/ui && bun run test:run`, `cd web/apps/dashboard && bun run test:run`.

## Verified Repo Constraints
- Go pre-commit only checks formatting on staged `*.go` files via `.githooks/pre-commit.d/01-go-fmt.sh`; run `gofmt -w` before committing Go changes.
- `.golangci.yml` has non-default constraints agents often miss:
  - `reflect` is denied by `depguard`.
  - `revive` enforces context-first params and `any` instead of `interface{}`.
  - `sloglint` requires lowercased static messages, snake_case keys, context-aware logging, and KV-only args.
- Go build tasks set `CGO_ENABLED=0` and `-ldflags='-s -w'`; keep that behavior when reproducing builds manually.
- `web/format` only formats `**/*.{ts,tsx,md}`. It will not touch JSON/YAML.

## CI / Change Detection
- `.github/workflows/build-on-push.yml` builds feature branches only for changed areas.
- Any `web/packages/**` change marks `console`, `dashboard`, and `website` as changed in CI.
- Changes under `api/**`, `cli/**`, `controllers/**`, `pkg/**`, `types/**`, `go.mod`, or `go.sum` trigger Go backend binary builds.
- `desktop` has its own change filter and build workflow path separate from the Next.js apps.

## Local Setup Reality
- `SETUP.md` is centered on `devenv/` and a local K3s stack, not just `go test` / `bun dev`.
- The documented bring-up order is: `task web:install` -> `docker-compose up k3s pre-app` -> `task api:dev` -> `docker-compose up post-app` -> `task web:dev`.

## Project Agent Workflow
- Engineering agents work from their own persistent worktree, not the shared checkout.
- Before editing, engineers switch their worktree to `development`, update it, then create a fresh `engineering/<role>/<task>` branch.
- PRs must come from `engineering/*` branches, never directly from `development`.
- After a PR is merged, switch that engineer's worktree back to `development` and update it before the next task.
- Engineers edit only their assigned area: API agents stay in `api/`, web agents stay in `web/`, Kubernetes agents stay in controller/operator/CRD/host-management areas plus generated manifests when regenerated.
- If a task needs cross-module changes, stop and route it through the coordinator instead of editing outside scope.
- Do not mutate deployed environments without explicit human approval. Read-only diagnosis such as `kubectl get`, `kubectl describe`, and logs is allowed with approval.

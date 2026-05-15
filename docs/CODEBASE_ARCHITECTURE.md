# Codebase Architecture

A per-section tour of the Kloudlite monorepo: where code lives, how it's built, how the pieces talk to each other. Complements `docs/RUNTIME_ARCHITECTURE.md` (which focuses on the deployment topology) and `README.md` (which focuses on product capabilities).

Monorepo root: `github.com/kloudlite/kloudlite`
- `api/` — Go 1.24 module (`github.com/kloudlite/kloudlite/api`). Backend, controllers, CLI binaries.
- `web/` — Bun 1.1 + Turbo workspace. Next.js apps and the Electron desktop app.
- `manifests/` — Generated CRD YAML. Do **not** hand-edit; run `task api:manifests`.
- `devenv/` — Local K3s bring-up.
- `supabase/` — Deno edge functions (website CMS).

---

## 1. Frontend (`web/`)

### Toolchain

- **Bun 1.1.40** package manager, **Turbo 2.3.3** task runner.
- Next.js **16.1.0**, React **19.1.0**, TypeScript **5**, Tailwind **4** (via `@tailwindcss/postcss`).
- Build output: `output: 'standalone'` for containerization.
- Testing: **Vitest 4** with `happy-dom`.

Root `web/package.json` drives per-app commands via Turbo filters: `bun run dev:console | dev:dashboard | dev:website | dev:desktop`, `build:*`, `lint`, `format`. `web/pnpm-workspace.yaml` declares the workspaces even though the package manager is Bun.

### Apps (`web/apps/`)

| App | Port | Role |
|---|---|---|
| `website` | 3000 | Marketing site, blog, docs, slides. `src/data/blog-posts.ts` drives `src/app/(main)/blog/[slug]/page.tsx`. Supabase CMS. |
| `dashboard` | 3001 | Core operational UI. Auth, environments, workspaces, admin, `kltun` VPN management. Server actions under `src/app/actions/` (`environment.actions.ts`, `composition.actions.ts`, etc.). |
| `console` | 3002 | Installation / cluster management portal. Auth-gated routes under `src/app/(authenticated)/`. Stripe + Turnstile integrations. |

All three apps share the `@/*` → `./src/*` path alias.

### Shared packages (`web/packages/`)

- **`@kloudlite/ui`** — ~50 components over Radix primitives + shadcn/ui conventions; `class-variance-authority` for variants, `react-hook-form` + `zod` for forms, `cn()` (via `clsx` + `tailwind-merge`) for className merging. Peer-depends on React 19 / Next 16.
- **`@kloudlite/lib`** — Root export is general utilities; `./k8s` sub-export (`src/k8s/index.ts`) exposes typed Kubernetes repositories used by server actions. Deps: `@kubernetes/client-node`, `@supabase/supabase-js`, `jose` (JWT).
- **`@kloudlite/types`** — Pure TypeScript types for composition / environment / machine / workspace / user / auth. No runtime deps.

Shared-package `lint` scripts are placeholders — don't rely on `turbo run lint` to validate package changes.

### Talking to the backend

The dashboard is **not** a thin client in front of a REST API. Server actions import `@kloudlite/lib/k8s` and speak directly to the Kubernetes API via `@kubernetes/client-node`; the Go `server` binary only runs admission webhooks and health endpoints (see Backend section). Auth is **NextAuth v5 beta**, JWT-based — `getSession()` is cached per request.

### Styling

Tailwind 4 via `postcss.config.ts` in every app. Design primitives in `@kloudlite/ui`; `next-themes` (website) for light/dark.

### Testing

`vitest.config.ts` per app, tests co-located next to source (`foo.test.tsx`). Coverage via v8. Run package-local, e.g. `cd web/packages/lib && bun run test:run`.

---

## 2. Backend (`api/`)

Go 1.24, module path `github.com/kloudlite/kloudlite/api`. **Stateless** — Kubernetes etcd is the source of truth for all domain state; an internal OCI registry stores snapshot artifacts.

### Binaries (`api/cmd/`)

Controller-section binaries are documented below; CLI binaries in the CLI section. The non-controller, non-CLI services are:

- **`server`** — Control-plane HTTPS server on `:8443`. Runs two things in the same process: (1) the controller-runtime manager (see Controllers), (2) a Gin router for Kubernetes admission webhooks + health. Entry: `api/cmd/server/main.go` → `internal/server/server.go`. Config via `kelseyhightower/envconfig` in `internal/config/config.go`. Installs `ValidatingWebhookConfiguration` / `MutatingWebhookConfiguration` on startup using an embedded `internal/services/webhook_configs.yaml`.
- **`tunnel-server`** — TLS WebSocket on `:443`, WireGuard peer relay. DNS proxy via `miekg/dns`. Watches Workspace/Environment CRDs with controller-runtime client.
- **`wm-ingress-controller`** — HTTP/HTTPS router (`:8000`/`:8443`) for per-environment hostnames. Reconciles core Kubernetes `Ingress` objects (not a CRD). Has its own `Taskfile.yml` include at the repo root.
- **`code-analyzer`** — Watches `/var/lib/kloudlite/home/workspaces` and serves analysis reports on `:8082` with a debounced (~45s) work queue; output at `/var/lib/kloudlite/code-analysis`.
- **`workmachine-node-manager`** — Host-level daemon on each WorkMachine node. Watches `PackageRequest` / `Snapshot` CRDs; manages the shared home filesystem at `/var/lib/kloudlite/home` (UID/GID 1001). Finalizers: `workspaces.kloudlite.io/directory-cleanup`, `workspaces.kloudlite.io/package-cleanup`.

### `internal/` layout

- **`controllers/`** — Domain reconcilers plus `manager.go` which wires the scheme and starts the controller-runtime manager. See the Controllers section.
- **`webhooks/`** — Gin handlers behind `/webhooks/{validate,mutate}/{kind}` for User, Environment, Workspace, WorkMachine, Snapshot/SnapshotRequest, MachineType, Pod/Service mutations, ConfigMap/Secret envvar injection.
- **`services/`** — `webhook_installer.go` installs the webhook configuration objects from the embedded YAML.
- **`middleware/`** — Zap request logging, CORS, correlation IDs.
- **`handlers/`** — `/health`, `/ready`.
- **`k8s/client.go`** — Wraps client-go `Clientset` + controller-runtime `client.WithWatch`, registers custom schemes, tunes QPS 50 / burst 100.
- **`config/config.go`** — `TLSConfig`, `AuthConfig` (JWT verify-only secret, registry RSA private key), `KubernetesConfig`, `InstallationConfig`, `RegistryConfig`.
- **`server/server.go`** — `Server` struct: wires TLS HTTP server, K8s client, controller manager, webhook router.

### `pkg/` (exported)

- **`logger/`** — Zap factory and a logger interface used by webhooks.
- **`oci/`** — Registry client (`google/go-containerregistry`). `Push()` writes BTRFS `send` streams as OCI layers with parent digest labels.
- **`imageref/`** — OCI reference parsing.
- **`operator-toolkit/`** — Generic reconciler plumbing (`reconciler/`), template engine with Sprig functions, step-based reconciliation patterns, event predicates.
- **`udptunnel/`** — WebSocket + UDP transport for WireGuard peer relay.
- **`utils/`** — Small helpers (sanitization, string ops).

### Persistence & integrations

- **No MongoDB/Postgres.** State = CRDs in etcd.
- **OCI registry** at `http://image-registry.kloudlite.svc.cluster.local:5000` (default) for snapshot artifacts.
- **Filesystem** state lives node-local on each WorkMachine (home, analysis output).
- Cloud SDKs: AWS, Azure, GCP, Oracle OCI — imported from the WorkMachine cloud provider package (see Controllers).
- **Compose** support: `compose-spec/compose-go/v2` converts Docker Compose into CRD specs.
- **Auth**: frontend mints tokens, backend verifies only. JWT via `golang-jwt/jwt/v5`.

### Lint / pre-commit

- `api/.golangci.yml` denies `reflect` via depguard.
- `revive` forces context-first parameters and `any` over `interface{}`.
- `sloglint` requires lowercase static messages, snake_case keys, KV-only args, context-aware logging.
- `.githooks/pre-commit.d/01-go-fmt.sh` runs `gofmt -w` on staged `api/**/*.go`.
- Build flags (Taskfile): `CGO_ENABLED=0 -ldflags='-s -w'`.

---

## 3. Controllers (`api/internal/controllers/`)

Built on **controller-runtime** (kubebuilder v3). All controllers share the single manager started by `cmd/server/main.go` via `controllers/manager.go`, which registers every custom scheme and sets up field indexes (e.g. by `spec.ownedBy`). Metrics/health on the manager are disabled because ports are shared with the webhook server; webhooks are installed by `services.WebhookInstaller`.

Generated CRD manifests are emitted to `manifests/` at repo root and re-generated via `task api:manifests`.

### Controllers and their CRDs

**`workspace/` — `workspaces.kloudlite.io/v1 Workspace`** (namespaced, owned by a WorkMachine)
- Spec: `displayName`, `ownedBy`, `workmachineName`, `environmentConnection`, `gitRepository`, `settings` (autostop, idleTimeout, envVars, startupScript), `resourceQuota`, `vsCodeVersion`, `fromSnapshot`, `expose[]` ports.
- Creates: VS Code server Pods in the WorkMachine's target namespace, ConfigMaps for git/SSH, NetworkPolicies, RoleBindings, a ServiceMonitor for idle detection.
- `lifecycle.go` manages pod startup/shutdown; `deletion.go` cleans up via finalizer `workspaces.kloudlite.io/finalizer`.

**`environment/` — `environments.kloudlite.io/v1 Environment`** (namespaced, owned by WorkMachine)
- Spec: auto-generated `targetNamespace` (`env-{name}-{random}`), `ownedBy`, `activated` (scales child Deployments/StatefulSets to 0 when false), `resourceQuotas`, `networkPolicies`, `compose`, `fromSnapshot`.
- Creates: the target namespace with quotas + NetworkPolicies; converts `Composition` (Docker Compose) to Kubernetes objects via `env_compose.go`.
- Sibling reconcilers in the same package: `EnvironmentSnapshotRequest`, `EnvironmentSnapshotRestore`, `EnvironmentForkRequest`.

**`workmachine/` — `machines.kloudlite.io/v1 WorkMachine`** (cluster-scoped)
- Spec: `displayName`, `ownedBy`, `targetNamespace`, `state` (running/stopped/disabled), `machineType` (EC2-style), `volumeSize` (50–1000 GB BTRFS, separate from 50 GB root), `autoShutdown`, `sshPublicKeys[]`, K3s join config (`k3sVersion`, `k3sServerURL`, `k3sAgentToken`).
- Creates: cloud VM (AWS/GCP/Azure/OCI via `cloud/` subdir), tunnel-server Deployment, SSH ConfigMap; registers the VM as a K3s node.
- Feature modules: `auto_shutdown.go`, `buildkit.go`, `ssh.go`, `machine_type.go`.
- Finalizer: `workmachine.machines.kloudlite.io/cleanup`.

**`snapshot/` — `snapshots.kloudlite.io/v1 {Snapshot, SnapshotRequest, SnapshotArtifact, SnapshotRestore}`**
- `SnapshotRequest` → BTRFS stream pushed to OCI registry → `Snapshot` materialized. `SnapshotReconciler` enforces `spec.retentionPolicy.expiresAt` and deletes expired artifacts from the registry.

**`user/` — `platform.kloudlite.io/v1alpha1 User`** (namespaced)
- Password hashing and SSH key validation happen in the mutating webhook, not in the controller. The controller handles finalizer cleanup (`user.platform.kloudlite.io/cleanup`) only.

**`packages/` — `packages.kloudlite.io/v1 PackageRequest`**
- Types live here; the reconciliation path runs in `workmachine-node-manager` (which holds the filesystem, not the control plane).

**`composition/`**
- Not an independent reconciler. `Composition` is embedded in `Environment.Spec.Compose`; `composition_parser.go` converts Docker Compose YAML into the CRD spec.

**`wmingress/`**
- Packaged alongside the others but started from the separate `cmd/wm-ingress-controller/` binary. Reconciles native Kubernetes `Ingress` resources (not a CRD), not the shared manager.

### Shared helpers

- `shared/pod_deletion_tracker.go` — de-conflicts pod deletion between the Workspace and WorkMachine controllers.
- `pkg/operator-toolkit/reconciler` — base `Reconciler` interface and unified `Status` subresource handling used by every controller above.

### Patterns worth knowing

- **Ownership:** Workspace and Environment are owned by WorkMachine; child Pods/ConfigMaps/RoleBindings are owned by their parent CRD. Cascading deletes through owner refs.
- **Finalizers:** Every controller owns one finalizer; deletion logic runs before finalizer removal.
- **Field indexes** on `spec.ownedBy` let list queries scope to a single WorkMachine without scans.

---

## 4. CLI Tools (`api/cmd/{kl,kli,kltun}`)

All three CLIs use **Cobra**. Versions are injected via `ldflags`. Commands are implemented one-file-per-command under each binary's `cmd/` subpackage.

### `kl` — in-workspace CLI

Runs **inside** the workspace container. Distributed as a binary baked into the workspace image, not a standalone download.

- Entry: `api/cmd/kl/main.go` → `cmd.Execute()` → `RootCmd` (`api/cmd/kl/cmd/root.go`).
- Commands: `status` (`st`), `pkg`/`package`, `config`/`cfg`, `env`/`environment`, `intercept`/`i`, `mcp`, `expose`, `image`, `completion`, `version`.
  - `pkg` hits Devbox's registry for Nix packages and creates `PackageRequest` CRDs.
  - `intercept` rewires environment service traffic back to the workspace.
  - `mcp` speaks Model Context Protocol (Claude / Cursor / OpenCode) against the workspace.
- Talks to the in-cluster Kubernetes API using controller-runtime client; reads CRDs from `github.com/kloudlite/kloudlite/api/internal/controllers/*/v1`. Kubeconfig comes from the service account or `$KUBECONFIG`.
- State is ephemeral: namespace/workspace come from `WORKSPACE_NAME` / `WORKSPACE_NAMESPACE` env vars; transient context cached at `/tmp/kloudlite-context.json`.

### `kli` — installer CLI

Cross-platform installer for deploying Kloudlite to a cloud account.

- Entry: `api/cmd/kli/main.go`.
- Commands: `aws`, `gcp`, `azure`/`az`, `oci`, each with `doctor` / `install` / `uninstall`.
- Uses cloud SDKs + direct kubectl/helm to provision infrastructure. Install manifests are embedded (`install_manifests.go`).
- Distribution: built per-platform via `Taskfile.yml` targets (`linux/darwin/windows × amd64/arm64`). Published as a GitHub Release (`_build-binary.yml` → `_release-binary.yml`) with SHA256 checksums.

### `kltun` — WireGuard tunnel client

Runs on the developer's laptop as a system daemon (launchd on macOS, systemd on Linux, service on Windows).

- Entry: `api/cmd/kltun/main.go`.
- Commands: `connect` (takes `--token`, `--server`), `install-ca` / `uninstall-ca` (system trust store), `quit`, `version`. Exposes a local HTTPS status endpoint.
- Implementation uses `pkg/api` (dashboard API client), `pkg/udptunnel` / `pkg/wireguard` / `pkg/wgkeys` (tunnel transport + key material), `pkg/hosts` (hosts-file DNS), `pkg/daemon` (service lifecycle over Unix socket / named pipe).
- On-disk state under `~/.kloudlite/`:
  - `deviceid` — persistent per-machine device ID
  - `wgkeys/` — WireGuard keys
  - Credentials are **not** persisted; the short-lived `--token` is exchanged for a session token kept only in memory.
- Distribution: native binary, same cross-platform build path as `kli`.

### CI for CLI binaries

`.github/workflows/build-on-push.yml` triggers CLI rebuilds on `api/cmd/kli/**` or `api/cmd/kltun/**` changes, independent of Docker image builds.

---

## 5. Desktop App (`web/apps/desktop/`)

**Electron** (v35), **not Tauri**. 100% TypeScript/JavaScript — no Rust.

### Build toolchain

- `electron-vite` (v5) for dev/build.
- `electron-builder` (v26) for packaging (`.dmg`/`.zip` / `.AppImage`/`.deb` / `.exe`).
- V8 bytecode caching via `v8-compile-cache`; manual chunk splitting for CodeMirror / React / icon libraries.

### Layout

- `src/main/` — Electron main process. `index.ts` creates the `BrowserWindow`; `tab-manager.ts` manages embedded `WebContentsView` tabs.
- `src/preload/` — context-bridge layer. `index.ts` exposes the typed `ElectronAPI` on `window.electronAPI`. `webview.ts` is a nested-webview preload.
- `src/renderer/` — React 19 UI.
  - `app.tsx` routes between three modes (`environments`, `workspaces`, `browse`) — switchable with `CmdOrCtrl+1/2/3`.
  - `components/` — `environment-content.tsx` (largest, ~28 KB), `workspace-content.tsx`, `webview-area.tsx`, `sidebar-*.tsx`, `services-graph/`, `code-editor.tsx`.
  - `store/` — Zustand stores: `tabs.ts`, `mode.ts`, `environments.ts`, `history.ts`.
  - `styles/` — Tailwind + CodeMirror theme overrides.

### Screens / flows

- **Environments** — list environments, inspect services (ClusterIP, DNS, target ports), show interception state (e.g. `interceptedBy: ws-1`).
- **Workspaces** — browse workspaces, manage intercepts.
- **Browse** — multi-tab webview area with built-in shortcuts (new tab, close tab, address bar, reload, sidebar toggle).
- Dark/light theme sync with the OS; TLS chain inspection via the `get-certificate` IPC (lazy-loads Node's `tls`, walks peer cert → issuers).

### IPC

Standard Electron pattern: `ipcMain.handle()` + `ipcRenderer.invoke()`. Handlers registered in `src/main/index.ts`:

- `window-control` — minimize/maximize/close
- `get-theme`, `theme-changed`
- `get-certificate` — TLS chain viewer
- `show-context-menu`, `show-popup-menu`
- `open-devtools`, `open-url-in-new-tab`
- `shortcut` — keyboard events routed from main into the renderer

The renderer does not call the Go API directly. For anything authenticated, it navigates an embedded `DashboardWebview` to `https://dashboard.kloudlite.io` (controlled by `DASHBOARD_BASE_URL`). Intercept state mirrors what `kltun` would surface; today it's seeded from a dummy store.

### CI / packaging

Build path is its own CI workflow at `.github/workflows/_build-desktop.yml`, independent of the Next.js app matrix. Build matrix: macOS x64/arm64, Linux amd64, Windows x64. Steps: `bun install` → `electron-vite build` (output to `out/`) → `electron-builder` → code-sign + notarize on macOS → GitHub Release.

---

## Cross-cutting notes

### Source of truth
| Concern | Lives in |
|---|---|
| Domain state (users, environments, workspaces, machines, snapshots) | Kubernetes etcd via CRDs |
| Snapshot artifacts (BTRFS streams) | In-cluster OCI registry |
| Workspace home / package data | Node-local filesystem on each WorkMachine |
| Marketing + blog content | Supabase (used by `web/apps/website`) |
| Auth | Frontend mints JWT (NextAuth); backend verifies only |

### Build & CI at a glance
- Root `Taskfile.yml` includes only `api` and `api/cmd/wm-ingress-controller`. Web commands run from `web/`.
- `build-on-push.yml` uses `dorny/paths-filter`: `api/**` → all Go images; `web/apps/<app>/**` or `web/packages/**` → that web app; `api/cmd/kli/**` / `api/cmd/kltun/**` → CLI binaries; `web/apps/desktop/**` → desktop build.
- Nightly (`build-nightly.yml`, 18:27 UTC) builds everything, publishes nightly release, auto-deploys dev via the cross-repo `kloudlite-ci/deploy.yml`.
- UAT: manual dispatch with per-app checkboxes. Master: `v*` tag push.

### Deploy repo
Deployments, e2e tests, and Helm charts live in [`kloudlite/kloudlite-ci`](https://github.com/kloudlite/kloudlite-ci). This repo **builds**; that repo **deploys**.

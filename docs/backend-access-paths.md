# Backend access paths

This note records the boundary between user-facing CLI flows, local Kubernetes controllers, and `platform-api`.

## `kl` workspace CLI

`cli/kl` is a workspace-local Kubernetes client. It should not know that `platform-api` exists.

- Reads and writes workspace, environment, composition, package, and intercept state directly through Kubernetes APIs.
- Uses in-cluster config first, then kubeconfig fallback, via controller-runtime/client-go clients.
- Must not import `api/*`, Connect RPC clients, or `platform-api` configuration.
- Commands such as `kl env status`, `kl env connect`, `kl env disconnect`, `kl intercept list`, `kl intercept start`, and `kl intercept stop` stay Kubernetes-direct.

## `platform-api`

`platform-api` is a separate HTTPS/Connect RPC entry point for API clients such as the desktop app.

- Serves resource RPC, REST resource adapters, and Kubernetes admission webhook routes.
- Owns API/runtime composition, CRD installation, resource registry/store/watch, and the fixed platform controller set.
- Does not replace `kl` workspace CLI Kubernetes access.

## `kltun`

`cli/kltun` may call tunnel/VPN HTTP APIs as part of tunnel setup. That path is separate from the `kl` workspace CLI and is not a reason for `cli/kl` to depend on `platform-api`.

## Current inventory

| User path | Primary access path | Notes |
| --- | --- | --- |
| `kl env status` | Kubernetes direct | Reads current Workspace status. |
| `kl env connect` | Kubernetes direct | Updates Workspace spec environment connection. |
| `kl env disconnect` | Kubernetes direct | Removes workspace-owned intercept specs, then clears Workspace environment connection. |
| `kl intercept list/status` | Kubernetes direct | Reads Environment/Composition status. |
| `kl intercept start/stop` | Kubernetes direct | Updates Environment/Composition intercept specs. |
| Desktop/resource clients | `platform-api` | Use HTTPS/Connect RPC resource APIs. |
| Admission webhooks | Kubernetes webhook HTTPS routes on `platform-api` | Remain webhook routes; not Connect RPC. |
| `kltun` VPN setup | Tunnel/VPN HTTP APIs | Separate tunnel API path. |

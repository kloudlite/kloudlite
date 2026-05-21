# Merge Host Manager Into WorkMachine Manager Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Run the WorkMachine machine-scoped controller and the current host-manager/node-manager controllers in one `workmachine-manager` pod per WorkMachine.

**Architecture:** Convert `cli/workmachine-node-manager` from a standalone `main` package into an importable runtime package, then start it alongside `controllers.NewMachineScopedManager` from `kloudlite server workmachine-manager`. Remove the separate `host-manager` StatefulSet and make the per-WorkMachine `workmachine-manager` StatefulSet carry the host privileges, mounts, env, and RBAC that host-manager used to require.

**Tech Stack:** Go 1.24, controller-runtime, Kubernetes RBAC/StatefulSet resources, Taskfile, Docker/BuildKit.

---

## File map

- Modify `cli/workmachine-node-manager/*.go`: change package from `main` to an importable package and expose a `Start(ctx, config)` entrypoint.
- Modify `cli/kloudlite/cmd/server_machine_scoped.go`: start both machine-scoped manager and host-manager runtime under one command.
- Modify `controllers/workmachine/machinescoped/controller.go`: remove `ensure-host-manager` as a managed child workload while keeping host-manager readiness represented by the single manager pod.
- Modify `controllers/workmachine/machinescoped/resources.go`: remove separate `host-manager` StatefulSet/Service creation or leave only cleanup helpers.
- Modify `controllers/workmachine/machinescoped/rbac.go`: bind host-manager RBAC to the per-WorkMachine `workmachine-manager-<name>` service account, not `host-manager`.
- Modify `controllers/workmachine/platformscoped/controller.go`: make the `workmachine-manager-<name>` StatefulSet privileged, `hostPID`, and host-mounted.
- Modify `build-sys/images/workmachine-manager/Dockerfile`: switch from distroless to a base containing host-manager tools.
- Test with focused Go package tests and dry-run image tasks.

## Task 1: Make host-manager importable

- [ ] Change `cli/workmachine-node-manager/*.go` from `package main` to `package workmachinenodemanager`.
- [ ] Replace `main()` in `cli/workmachine-node-manager/main.go` with `Start(ctx context.Context, cfg Config) error`.
- [ ] Define `Config` with `Namespace`, `WorkMachineName`, `SnapshotRegistryEndpoint`, `SnapshotRegistryPrefix`, and `SnapshotRegistryInsecure`.
- [ ] Keep existing setup order: workspace home, SSH config dir, schemes, controller-runtime manager, package reconciler, SSH, workspace cleanup, GPU, snapshot request/restore, storage GC.
- [ ] Run: `go test -count=1 ./cli/workmachine-node-manager`.
- [ ] Expected: package compiles and tests pass.

## Task 2: Start both runtimes from `workmachine-manager`

- [ ] In `cli/kloudlite/cmd/server_machine_scoped.go`, start `controllers.NewMachineScopedManager` as today.
- [ ] Import the new host-manager runtime package.
- [ ] Start host-manager runtime in a goroutine using env/config values loaded from existing `config.Load()`.
- [ ] Return if either runtime exits with an error, or shutdown on context/signal.
- [ ] Run: `go test -count=1 ./cli/kloudlite/cmd`.
- [ ] Expected: command tests pass.

## Task 3: Stop creating a separate host-manager workload

- [ ] In `controllers/workmachine/machinescoped/controller.go`, remove the `ensure-host-manager` lifecycle step that creates/deletes `StatefulSet/host-manager`.
- [ ] Keep cleanup logic for existing `host-manager` StatefulSet/Service so upgrades remove old workloads.
- [ ] Mark `ConditionHostManagerReady` true when old workload cleanup completes and the integrated runtime is assumed present in the current pod.
- [ ] Run: `go test -count=1 ./controllers/workmachine/machinescoped`.
- [ ] Expected: tests pass.

## Task 4: Move host-manager RBAC to workmachine-manager service account

- [ ] In `controllers/workmachine/machinescoped/rbac.go`, change host-manager RBAC subject from `ServiceAccount/host-manager` in target namespace to `ServiceAccount/workmachine-manager-<workmachine>` in the platform namespace used by the platform-scoped manager StatefulSet.
- [ ] Preserve the same host-manager permissions for PackageRequests, Snapshots, SnapshotRequests, SnapshotRestores, Workspaces, Nodes, Environments, and Secrets.
- [ ] Keep cleanup or overwrite of old `host-manager` RoleBinding/ClusterRoleBinding so upgrades converge.
- [ ] Run: `go test -count=1 ./controllers/workmachine/machinescoped`.
- [ ] Expected: tests pass.

## Task 5: Make workmachine-manager pod host-capable

- [ ] In `controllers/workmachine/platformscoped/controller.go`, update `StatefulSet/workmachine-manager-<workmachine>` pod spec with `HostPID: true`.
- [ ] Add privileged `SecurityContext` to the main container.
- [ ] Add env: `NAMESPACE`, `WORKMACHINE_NAME`, `SNAPSHOT_REGISTRY_ENDPOINT`, `SNAPSHOT_REGISTRY_PREFIX`, `SNAPSHOT_REGISTRY_INSECURE`, and `NODE_NAME`.
- [ ] Add host mounts copied from old host-manager: `/var/lib/kloudlite`, `/sys`, `/dev`, `/proc`, `/lib/modules`, and Nix store handling.
- [ ] Run: `go test -count=1 ./controllers/workmachine/platformscoped ./controllers/workmachine/machinescoped`.
- [ ] Expected: tests pass.

## Task 6: Update workmachine-manager image

- [ ] Replace `build-sys/images/workmachine-manager/Dockerfile` distroless base with a Debian/Ubuntu base that includes `bash`, `coreutils`, `util-linux`, `btrfs-progs`, `curl`, `xz-utils`, `git`, and `ca-certificates`.
- [ ] Install Nix prerequisites or copy/install Nix consistently with the old host-manager image expectations.
- [ ] Keep `ENTRYPOINT ["/app/workmachine-manager"]` and `CMD ["server", "workmachine-manager"]`.
- [ ] Run: `task --dry build-sys:image:workmachine-manager`.
- [ ] Expected: dry-run shows Dockerfile path and tag resolution.

## Task 7: End-to-end verification

- [ ] Run: `go test -count=1 ./cli/workmachine-node-manager ./cli/kloudlite/cmd ./controllers/workmachine/platformscoped ./controllers/workmachine/machinescoped ./controllers`.
- [ ] Run: `task --dry build-sys:image:workmachine-manager`.
- [ ] Run: `git diff --check`.
- [ ] Expected: all commands exit 0.

## Self-review

- Spec coverage: covers single pod/process, host-manager controllers, RBAC, pod privileges/mounts, image tooling, cleanup of old separate workload.
- Placeholder scan: no open TBD/TODO placeholders.
- Type consistency: names match current files and current runtime naming.

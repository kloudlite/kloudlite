# Workspace Controller in WorkMachine Manager Design

## Goal

Move Workspace reconciliation into the existing per-WorkMachine `workmachine-manager` runtime so Workspace becomes a machine-scoped controller, matching the resource model where each Workspace belongs to exactly one WorkMachine via `Workspace.spec.workmachine`.

## Current State

- `workmachine-manager` currently runs the machine-scoped WorkMachine controller, scoped Environment controller, integrated `wm-ingress`, and workmachine-node-manager runtime.
- `WorkspaceReconciler` is currently registered from the non-platform-only API/controller manager path.
- Workspace objects already declare their WorkMachine through `spec.workmachine` (`WorkspaceSpec.WorkmachineName`).
- Workspace helper code resolves target namespace by reading the referenced WorkMachine and using `WorkMachine.spec.targetNamespace`.

## Desired Runtime Ownership

- The only long-lived Workspace reconciler should run inside `workmachine-manager`.
- Platform/API controller managers should not reconcile Workspaces after this migration.
- A `workmachine-manager` instance must reconcile only Workspace resources belonging to its own WorkMachine.

## Scoping Rules

`WorkspaceReconciler` will gain explicit runtime scope fields:

- `OwnNamespace`: sourced from `POD_NAMESPACE`, falling back to `NAMESPACE`.
- `WorkMachineName`: sourced from `WORKMACHINE_NAME`.

When scope is configured, the reconciler must process a Workspace only if both are true:

- `workspace.Namespace == OwnNamespace`
- `workspace.Spec.WorkmachineName == WorkMachineName`

If either `OwnNamespace` or `WorkMachineName` is missing while the reconciler is used in machine-scoped mode, it fails closed and skips all Workspace reconciliation.

For compatibility with existing unit tests and any legacy all-in-one manager construction paths during refactor, an entirely unscoped `WorkspaceReconciler` remains permissive. Partial scope is never permissive.

## WorkMachine Lookup Guard

Workspace helper paths that fetch a WorkMachine must respect the same machine boundary. In a scoped reconciler, requests for any WorkMachine name other than `WorkMachineName` return an error before querying the API server. This prevents a Workspace in the current namespace from accidentally deriving runtime configuration from a different WorkMachine.

## Controller Registration

`newMachineScopedManager` will instantiate and register `workspace.WorkspaceReconciler` with:

- controller-runtime client and scheme from the manager,
- REST config and Kubernetes clientset,
- controller config from `LoadConfig`,
- scope fields from the runtime environment.

The normal API manager path will stop registering `WorkspaceReconciler`. The platform-only API manager path already does not register Workspace and remains unchanged.

Machine-scoped controller names will include `workspace` for tests/logging, alongside `workmachine-manager` and `environment`.

## Event Filtering

The primary safety boundary is in `Reconcile` after fetching the Workspace: out-of-scope Workspaces return success without mutation. This avoids acting on unrelated resources even if watches enqueue them.

Secondary event handlers that map Environment, ClusterRole, or ClusterRoleBinding changes back to Workspace requests should filter to the current scope where practical. The most important path is `findWorkspacesForEnvironment`, which should only enqueue Workspaces matching `OwnNamespace` and `WorkMachineName` when scoped.

## Testing Strategy

Add focused tests before implementation:

- `WorkspaceReconciler` allows a Workspace whose namespace and `spec.workmachine` match scope.
- It skips Workspaces in a different namespace.
- It skips Workspaces for another WorkMachine.
- It fails closed when only one scope field is configured.
- It remains permissive when entirely unscoped for legacy/unit-test compatibility.
- Scoped `getWorkMachine` rejects a different WorkMachine name.
- Machine-scoped controller names include `workspace`.

Then update existing Workspace and controller tests as needed, without broad refactors.

## Runtime Verification

Local verification:

- `go test ./controllers/workspace ./controllers -v`
- `go test ./...`

Live dev-local verification after image build/restart:

- rebuild and push `workmachine-manager:dev-local`,
- restart `statefulset/workmachine-manager-<name>` in the WorkMachine namespace,
- verify the pod is running the new image digest,
- create or update a Workspace for the current WorkMachine and confirm the reconciler acts only on that Workspace,
- confirm a Workspace with a different `spec.workmachine` is skipped and not mutated by this manager.

## Deployment Impact

This changes where Workspace reconciliation runs. It requires rolling out the `workmachine-manager` image. It should not require API server changes for Workspace reconciliation, except that platform/API no longer owns that controller path. Existing Workspace CRDs and API resources remain unchanged.

## Out of Scope

- Renaming `workmachine-manager`.
- Introducing a separate `workspace-manager` runtime.
- Changing the Workspace CRD shape.
- Refactoring Workspace lifecycle/status semantics beyond the scope guard needed for machine ownership.

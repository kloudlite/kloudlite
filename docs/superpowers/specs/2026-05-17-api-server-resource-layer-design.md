# API Server Resource Layer Design

Date: 2026-05-17

## Summary

Add a resource-serving subsystem to `kloudlite server api`. The API server will list and watch selected Kloudlite Kubernetes resources, keep an in-memory cache, and expose typed HTTP APIs for read and write operations. This spec is limited to the Go API server. It does not migrate the dashboard or change dashboard code.

## Goals

- Make the Go API server the owner of a typed resource API surface.
- Serve reads from an in-memory cache populated by Kubernetes list/watch.
- Perform writes against Kubernetes and let watch events update the cache.
- Keep the API server package boundaries small and testable.
- Provide a stable server-side contract that dashboard or other clients can adopt later.

## Non-Goals

- No dashboard migration in this phase.
- No redesign of deployment manifests or installers.
- No new standalone controller or resource-server binary.
- No optimistic cache mutations after writes.
- No hand edits to generated manifests.

## Architecture

The resource layer runs inside the existing `kloudlite server api` process. It starts with the API server runtime and uses the same Kubernetes access model as the API server.

The external API is typed/domain-oriented. Internally, resource registration, watching, and storage are generic enough to avoid duplicating list/watch/store code for every custom resource.

Package boundaries:

- `api/resources/registry`
  - Defines resource metadata: alias, group, version, kind, plural, scope, object type, and list type.
  - Validates route/resource compatibility, such as namespaced routes for namespaced resources only.
- `api/resources/store`
  - Owns the thread-safe in-memory resource cache.
  - Stores objects by resource alias, scope, namespace, and name.
  - Maintains indexes for label lookups and `kloudlite.io/hash` where present.
- `api/resources/watch`
  - Owns list-then-watch lifecycle per registered resource and namespace scope.
  - Handles watch restarts, relists, and readiness state.
- `api/resources/service`
  - Provides the handler-facing operations: list, get, create, patch, and delete.
  - Reads from the store and writes to Kubernetes.
- `api/resources/handlers`
  - Registers HTTP routes.
  - Validates path parameters and request bodies.
  - Maps service errors to HTTP responses.

## Initial Resource Scope

The initial registry should cover resources already needed by the server-side resource API. The implementation plan can split these into smaller batches.

Cluster-scoped resources:

- `users`
- `userpreferences`
- `workmachines`
- `machinetypes`

Namespaced resources:

- `workspaces`
- `environments`
- `snapshots`
- `packagerequests`
- `services`
- `configmaps`
- `secrets`

Resource aliases are API-owned names, not raw Kubernetes discovery strings. The registry maps each alias to Kubernetes group/version/kind/plural and scope.

## API Surface

Routes are mounted on the existing API server. The exact prefix should follow existing API server routing conventions during implementation. The resource path shape is:

Cluster-scoped resources:

- `GET /resources/:resource`
- `GET /resources/:resource/:name`
- `POST /resources/:resource`
- `PATCH /resources/:resource/:name`
- `DELETE /resources/:resource/:name`

Namespaced resources:

- `GET /namespaces/:namespace/resources/:resource`
- `GET /namespaces/:namespace/resources/:resource/:name`
- `POST /namespaces/:namespace/resources/:resource`
- `PATCH /namespaces/:namespace/resources/:resource/:name`
- `DELETE /namespaces/:namespace/resources/:resource/:name`

Handlers validate:

- The resource alias exists.
- The requested route matches the resource scope.
- Namespace and name path parameters are non-empty and Kubernetes-compatible.
- Request body `apiVersion` and `kind`, when provided, match the registered resource.
- Namespaced write bodies target the namespace from the path.

Responses should return Kubernetes-style resource objects for get/create/patch results and lists for list operations. Errors should use the API server's existing response style where possible.

## Cache and Watch Behavior

Each watched resource follows list-then-watch:

1. List current objects from Kubernetes.
2. Populate the in-memory store.
3. Mark the resource scope as ready.
4. Start watching from the list `resourceVersion`.
5. Apply `ADDED`, `MODIFIED`, and `DELETED` events to the store.
6. On watch errors or expired resource versions, relist and restart the watch.

Read behavior:

- Reads are served from memory after the relevant cache scope is ready.
- If the cache scope is not ready, the service returns a clear cache-not-ready error mapped to HTTP `503`.
- Direct Kubernetes fallback for reads is not part of the first design because it can hide cache readiness and consistency issues.

Write behavior:

- Creates, patches, and deletes go to Kubernetes through the service layer.
- The store is not updated optimistically.
- The cache reflects writes only after Kubernetes emits watch events.
- Kubernetes errors are mapped to HTTP responses.

Readiness behavior:

- Cluster-scoped resources start watching during API server startup.
- Namespaced resources can be started lazily for requested namespaces or eagerly for known namespaces if that information already exists in the API server.
- A request for a namespace/resource pair that is not ready may trigger watch initialization and then return `503` until the first list completes.

## Store Indexes

The first store implementation should support:

- Resource alias + name for cluster-scoped objects.
- Resource alias + namespace + name for namespaced objects.
- Label indexes for common label selector reads.
- `kloudlite.io/hash` index where present.

Domain-specific indexes, such as user email or owner namespace, should be added only when handlers need them.

## Error Handling

The service layer should return typed errors that handlers can map consistently:

- Unknown resource alias: `404`.
- Wrong route scope for resource: `400`.
- Cache not ready or stale: `503`.
- Kubernetes not found: `404`.
- Kubernetes conflict: `409`.
- Kubernetes validation or admission error: `400`.
- Unexpected Kubernetes/client/server error: `500`.

Watch failures should be logged with enough resource and namespace context to diagnose which cache is affected. Temporary watch failures should not crash the API server; the watcher should relist and retry.

## Testing Strategy

Unit tests:

- Registry validates resource metadata and scope matching.
- Store applies add/update/delete events and maintains indexes.
- Store readiness gates reads.
- Service reads from store and writes through a fake Kubernetes client.
- Handlers map path/body validation and service errors to HTTP responses.

Integration-style tests with fake clients:

- List-then-watch startup populates readiness and store contents.
- Watch restart after an expired resource version relists and remains usable.
- Writes do not mutate the cache before a watch event arrives.

Verification commands for implementation should be focused first, then broadened:

- `go test ./api/resources/...`
- Relevant API server package tests.
- `go test ./...` before marking the implementation complete.

## Implementation Notes

- Prefer existing Kubernetes client setup in the API server rather than creating a new config path.
- Keep CRD type usage in the top-level `types/` packages shared by API and controllers.
- Do not add a root-level `internal/` package.
- Do not edit generated `manifests/` by hand.
- Keep the initial endpoint implementation small; broader resource-specific domain methods can be layered on later.

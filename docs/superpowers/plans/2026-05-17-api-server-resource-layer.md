# API Server Resource Layer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an API-server-only resource layer that lists/watches registered Kubernetes resources, serves reads from an in-memory cache, and writes through Kubernetes.

**Architecture:** Add focused packages under `api/resources`: `registry`, `store`, `service`, `watch`, and `handlers`. Keep the external HTTP API generic but typed by registered resource aliases. Implement reads against the store and writes against the existing controller-runtime `client.WithWatch` from `api/k8s`.

**Tech Stack:** Go, Gin, controller-runtime client, Kubernetes API machinery, existing Kloudlite CRD types under `types/`.

---

## File Structure

- Create `api/resources/registry/registry.go`: resource metadata, scope enum, registry lookup, default resource registrations.
- Create `api/resources/registry/registry_test.go`: alias lookup and scope validation tests.
- Create `api/resources/store/store.go`: thread-safe cache, key helpers, readiness, list/get/upsert/delete, basic label/hash filtering.
- Create `api/resources/store/store_test.go`: cache behavior and readiness tests.
- Create `api/resources/service/errors.go`: typed errors and HTTP status mapping helpers.
- Create `api/resources/service/service.go`: resource list/get/create/patch/delete operations over store and Kubernetes client.
- Create `api/resources/service/service_test.go`: fake client tests for reads/writes and no optimistic mutation.
- Create `api/resources/watch/watcher.go`: list-then-watch runner per resource/scope.
- Create `api/resources/watch/watcher_test.go`: fake watch tests for initial list and event application.
- Create `api/resources/handlers/handlers.go`: Gin route registration and request/response handling.
- Create `api/resources/handlers/handlers_test.go`: HTTP routing, validation, and status tests.
- Modify `api/server/webhook_routes.go`: register resource routes under `/api/v1`.
- Modify `api/server/server.go`: construct resource service/watch manager and start it with the API server.

---

### Task 1: Resource Registry

**Files:**
- Create: `api/resources/registry/registry.go`
- Test: `api/resources/registry/registry_test.go`

- [ ] **Step 1: Write failing registry tests**

Create `api/resources/registry/registry_test.go`:

```go
package registry

import "testing"

func TestDefaultRegistryFindsKnownResources(t *testing.T) {
	r := Default()

	workspace, ok := r.Get("workspaces")
	if !ok {
		t.Fatal("expected workspaces to be registered")
	}
	if workspace.Scope != Namespaced {
		t.Fatalf("expected workspaces to be namespaced, got %s", workspace.Scope)
	}
	if workspace.Kind != "Workspace" || workspace.Plural != "workspaces" {
		t.Fatalf("unexpected workspace metadata: %#v", workspace)
	}

	workMachine, ok := r.Get("workmachines")
	if !ok {
		t.Fatal("expected workmachines to be registered")
	}
	if workMachine.Scope != Cluster {
		t.Fatalf("expected workmachines to be cluster-scoped, got %s", workMachine.Scope)
	}
}

func TestRegistryRejectsDuplicateAliases(t *testing.T) {
	_, err := New([]Resource{
		{Alias: "users", Scope: Cluster},
		{Alias: "users", Scope: Cluster},
	})
	if err == nil {
		t.Fatal("expected duplicate alias error")
	}
}

func TestResourceScopeValidation(t *testing.T) {
	r := Default()

	if err := r.RequireScope("workspaces", Namespaced); err != nil {
		t.Fatalf("expected namespaced workspaces route to be valid: %v", err)
	}
	if err := r.RequireScope("workspaces", Cluster); err == nil {
		t.Fatal("expected cluster route for workspaces to be invalid")
	}
	if err := r.RequireScope("does-not-exist", Cluster); err == nil {
		t.Fatal("expected unknown resource to be invalid")
	}
}
```

- [ ] **Step 2: Run registry tests and verify they fail**

Run: `go test ./api/resources/registry -run Test -v`

Expected: FAIL because package/files do not exist or symbols are undefined.

- [ ] **Step 3: Implement registry**

Create `api/resources/registry/registry.go`:

```go
package registry

import (
	"fmt"
	"strings"

	environmentsv1 "github.com/kloudlite/kloudlite/types/environment/v1"
	packagesv1 "github.com/kloudlite/kloudlite/types/packages/v1"
	snapshotv1 "github.com/kloudlite/kloudlite/types/snapshot/v1"
	platformv1alpha1 "github.com/kloudlite/kloudlite/types/user/v1alpha1"
	machinesv1 "github.com/kloudlite/kloudlite/types/workmachine/v1"
	workspacesv1 "github.com/kloudlite/kloudlite/types/workspace/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Scope string

const (
	Cluster    Scope = "cluster"
	Namespaced Scope = "namespaced"
)

type Resource struct {
	Alias     string
	Group     string
	Version   string
	Kind      string
	Plural    string
	Scope     Scope
	NewObject func() runtime.Object
	NewList   func() runtime.Object
}

func (r Resource) GroupVersionKind() schema.GroupVersionKind {
	return schema.GroupVersionKind{Group: r.Group, Version: r.Version, Kind: r.Kind}
}

type Registry struct {
	resources map[string]Resource
}

func New(resources []Resource) (*Registry, error) {
	reg := &Registry{resources: map[string]Resource{}}
	for _, resource := range resources {
		resource.Alias = strings.TrimSpace(resource.Alias)
		if resource.Alias == "" {
			return nil, fmt.Errorf("resource alias is required")
		}
		if _, ok := reg.resources[resource.Alias]; ok {
			return nil, fmt.Errorf("resource alias %q is registered more than once", resource.Alias)
		}
		if resource.Scope != Cluster && resource.Scope != Namespaced {
			return nil, fmt.Errorf("resource %q has invalid scope %q", resource.Alias, resource.Scope)
		}
		reg.resources[resource.Alias] = resource
	}
	return reg, nil
}

func Default() *Registry {
	reg, err := New([]Resource{
		{Alias: "users", Group: platformv1alpha1.GroupVersion.Group, Version: platformv1alpha1.GroupVersion.Version, Kind: "User", Plural: "users", Scope: Cluster, NewObject: func() runtime.Object { return &platformv1alpha1.User{} }, NewList: func() runtime.Object { return &platformv1alpha1.UserList{} }},
		{Alias: "userpreferences", Group: platformv1alpha1.GroupVersion.Group, Version: platformv1alpha1.GroupVersion.Version, Kind: "UserPreferences", Plural: "userpreferences", Scope: Cluster, NewObject: func() runtime.Object { return &platformv1alpha1.UserPreferences{} }, NewList: func() runtime.Object { return &platformv1alpha1.UserPreferencesList{} }},
		{Alias: "workmachines", Group: machinesv1.GroupVersion.Group, Version: machinesv1.GroupVersion.Version, Kind: "WorkMachine", Plural: "workmachines", Scope: Cluster, NewObject: func() runtime.Object { return &machinesv1.WorkMachine{} }, NewList: func() runtime.Object { return &machinesv1.WorkMachineList{} }},
		{Alias: "machinetypes", Group: machinesv1.GroupVersion.Group, Version: machinesv1.GroupVersion.Version, Kind: "MachineType", Plural: "machinetypes", Scope: Cluster, NewObject: func() runtime.Object { return &machinesv1.MachineType{} }, NewList: func() runtime.Object { return &machinesv1.MachineTypeList{} }},
		{Alias: "workspaces", Group: workspacesv1.GroupVersion.Group, Version: workspacesv1.GroupVersion.Version, Kind: "Workspace", Plural: "workspaces", Scope: Namespaced, NewObject: func() runtime.Object { return &workspacesv1.Workspace{} }, NewList: func() runtime.Object { return &workspacesv1.WorkspaceList{} }},
		{Alias: "environments", Group: environmentsv1.SchemeGroupVersion.Group, Version: environmentsv1.SchemeGroupVersion.Version, Kind: "Environment", Plural: "environments", Scope: Namespaced, NewObject: func() runtime.Object { return &environmentsv1.Environment{} }, NewList: func() runtime.Object { return &environmentsv1.EnvironmentList{} }},
		{Alias: "snapshots", Group: snapshotv1.SchemeGroupVersion.Group, Version: snapshotv1.SchemeGroupVersion.Version, Kind: "Snapshot", Plural: "snapshots", Scope: Namespaced, NewObject: func() runtime.Object { return &snapshotv1.Snapshot{} }, NewList: func() runtime.Object { return &snapshotv1.SnapshotList{} }},
		{Alias: "packagerequests", Group: packagesv1.GroupVersion.Group, Version: packagesv1.GroupVersion.Version, Kind: "PackageRequest", Plural: "packagerequests", Scope: Namespaced, NewObject: func() runtime.Object { return &packagesv1.PackageRequest{} }, NewList: func() runtime.Object { return &packagesv1.PackageRequestList{} }},
		{Alias: "services", Group: "", Version: "v1", Kind: "Service", Plural: "services", Scope: Namespaced, NewObject: func() runtime.Object { return &corev1.Service{} }, NewList: func() runtime.Object { return &corev1.ServiceList{} }},
		{Alias: "configmaps", Group: "", Version: "v1", Kind: "ConfigMap", Plural: "configmaps", Scope: Namespaced, NewObject: func() runtime.Object { return &corev1.ConfigMap{} }, NewList: func() runtime.Object { return &corev1.ConfigMapList{} }},
		{Alias: "secrets", Group: "", Version: "v1", Kind: "Secret", Plural: "secrets", Scope: Namespaced, NewObject: func() runtime.Object { return &corev1.Secret{} }, NewList: func() runtime.Object { return &corev1.SecretList{} }},
	})
	if err != nil {
		panic(err)
	}
	return reg
}

func (r *Registry) Get(alias string) (Resource, bool) {
	resource, ok := r.resources[alias]
	return resource, ok
}

func (r *Registry) RequireScope(alias string, scope Scope) error {
	resource, ok := r.Get(alias)
	if !ok {
		return fmt.Errorf("unknown resource %q", alias)
	}
	if resource.Scope != scope {
		return fmt.Errorf("resource %q is %s scoped, not %s scoped", alias, resource.Scope, scope)
	}
	return nil
}

func (r *Registry) All() []Resource {
	resources := make([]Resource, 0, len(r.resources))
	for _, resource := range r.resources {
		resources = append(resources, resource)
	}
	return resources
}
```

- [ ] **Step 4: Run registry tests and verify they pass**

Run: `go test ./api/resources/registry -run Test -v`

Expected: PASS.

- [ ] **Step 5: Commit registry**

Run:

```bash
git add api/resources/registry
git commit -m "feat: add api resource registry"
```

---

### Task 2: In-Memory Store

**Files:**
- Create: `api/resources/store/store.go`
- Test: `api/resources/store/store_test.go`

- [ ] **Step 1: Write failing store tests**

Create `api/resources/store/store_test.go` with tests for readiness, get/list, label filtering, hash filtering, and delete:

```go
package store

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestStoreReadinessGatesReads(t *testing.T) {
	s := New()
	if s.Ready("configmaps", "default") {
		t.Fatal("expected scope to start not ready")
	}
	s.SetReady("configmaps", "default", true)
	if !s.Ready("configmaps", "default") {
		t.Fatal("expected scope to be ready")
	}
}

func TestStoreUpsertGetListAndDelete(t *testing.T) {
	s := New()
	cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "default", Labels: map[string]string{"app": "demo", "kloudlite.io/hash": "abc"}}}

	s.Upsert("configmaps", cm)
	got, ok := s.Get("configmaps", "default", "app")
	if !ok || got.GetName() != "app" {
		t.Fatalf("expected to get stored configmap, ok=%v got=%#v", ok, got)
	}

	items := s.List("configmaps", "default", Selector{Labels: map[string]string{"app": "demo"}})
	if len(items) != 1 {
		t.Fatalf("expected one item by label, got %d", len(items))
	}

	items = s.List("configmaps", "default", Selector{Hash: "abc"})
	if len(items) != 1 {
		t.Fatalf("expected one item by hash, got %d", len(items))
	}

	s.Delete("configmaps", "default", "app")
	if _, ok := s.Get("configmaps", "default", "app"); ok {
		t.Fatal("expected deleted object to be missing")
	}
}
```

- [ ] **Step 2: Run store tests and verify they fail**

Run: `go test ./api/resources/store -run Test -v`

Expected: FAIL because store is not implemented.

- [ ] **Step 3: Implement store**

Create `api/resources/store/store.go` with a mutex-protected map. Store `client.Object` values as `DeepCopyObject().(client.Object)` to prevent callers mutating cached values. Implement:

```go
package store

import (
	"sync"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Selector struct {
	Labels map[string]string
	Hash   string
}

type Store struct {
	mu      sync.RWMutex
	objects map[string]map[string]client.Object
	ready   map[string]bool
}

func New() *Store {
	return &Store{objects: map[string]map[string]client.Object{}, ready: map[string]bool{}}
}

func (s *Store) SetReady(alias string, namespace string, ready bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ready[scopeKey(alias, namespace)] = ready
}

func (s *Store) Ready(alias string, namespace string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ready[scopeKey(alias, namespace)]
}

func (s *Store) Upsert(alias string, object client.Object) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := scopeKey(alias, object.GetNamespace())
	if s.objects[key] == nil {
		s.objects[key] = map[string]client.Object{}
	}
	s.objects[key][object.GetName()] = copyObject(object)
}

func (s *Store) ReplaceScope(alias string, namespace string, objects []client.Object) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := scopeKey(alias, namespace)
	s.objects[key] = map[string]client.Object{}
	for _, object := range objects {
		s.objects[key][object.GetName()] = copyObject(object)
	}
	s.ready[key] = true
}

func (s *Store) Get(alias string, namespace string, name string) (client.Object, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	object, ok := s.objects[scopeKey(alias, namespace)][name]
	if !ok {
		return nil, false
	}
	return copyObject(object), true
}

func (s *Store) List(alias string, namespace string, selector Selector) []client.Object {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := []client.Object{}
	for _, object := range s.objects[scopeKey(alias, namespace)] {
		if matches(object, selector) {
			items = append(items, copyObject(object))
		}
	}
	return items
}

func (s *Store) Delete(alias string, namespace string, name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.objects[scopeKey(alias, namespace)], name)
}

func scopeKey(alias string, namespace string) string { return alias + "/" + namespace }

func copyObject(object client.Object) client.Object {
	return object.DeepCopyObject().(client.Object)
}

func matches(object client.Object, selector Selector) bool {
	if selector.Hash != "" && object.GetLabels()["kloudlite.io/hash"] != selector.Hash {
		return false
	}
	for key, value := range selector.Labels {
		if object.GetLabels()[key] != value {
			return false
		}
	}
	return true
}
```

- [ ] **Step 4: Run store tests and verify they pass**

Run: `go test ./api/resources/store -run Test -v`

Expected: PASS.

- [ ] **Step 5: Commit store**

Run:

```bash
git add api/resources/store
git commit -m "feat: add api resource store"
```

---

### Task 3: Resource Service and Errors

**Files:**
- Create: `api/resources/service/errors.go`
- Create: `api/resources/service/service.go`
- Test: `api/resources/service/service_test.go`

- [ ] **Step 1: Write failing service tests**

Create `api/resources/service/service_test.go` to cover cache reads and Kubernetes writes:

```go
package service

import (
	"context"
	"testing"

	"github.com/kloudlite/kloudlite/api/resources/registry"
	"github.com/kloudlite/kloudlite/api/resources/store"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestListRequiresReadyCache(t *testing.T) {
	svc := New(registry.Default(), store.New(), fake.NewClientBuilder().Build())
	_, err := svc.List(context.Background(), "configmaps", "default", store.Selector{})
	if !IsKind(err, ErrCacheNotReady) {
		t.Fatalf("expected cache not ready error, got %v", err)
	}
}

func TestCreateWritesToKubernetesWithoutOptimisticStoreMutation(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	st := store.New()
	svc := New(registry.Default(), st, fake.NewClientBuilder().WithScheme(scheme).Build())
	cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "default"}}

	created, err := svc.Create(context.Background(), "configmaps", "default", cm)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if created.GetName() != "app" {
		t.Fatalf("unexpected created object: %#v", created)
	}
	if _, ok := st.Get("configmaps", "default", "app"); ok {
		t.Fatal("store should not be updated optimistically")
	}
}
```

- [ ] **Step 2: Run service tests and verify they fail**

Run: `go test ./api/resources/service -run Test -v`

Expected: FAIL because service is not implemented.

- [ ] **Step 3: Implement service errors**

Create `api/resources/service/errors.go` with:

```go
package service

import (
	"errors"
	"fmt"
	"net/http"
)

type ErrorKind string

const (
	ErrUnknownResource ErrorKind = "unknown_resource"
	ErrWrongScope      ErrorKind = "wrong_scope"
	ErrCacheNotReady   ErrorKind = "cache_not_ready"
	ErrBadRequest      ErrorKind = "bad_request"
)

type Error struct {
	Kind    ErrorKind
	Message string
	Err     error
}

func (e *Error) Error() string { return e.Message }
func (e *Error) Unwrap() error { return e.Err }

func NewError(kind ErrorKind, message string, err error) error {
	return &Error{Kind: kind, Message: message, Err: err}
}

func IsKind(err error, kind ErrorKind) bool {
	var serviceErr *Error
	return errors.As(err, &serviceErr) && serviceErr.Kind == kind
}

func UnknownResource(alias string) error {
	return NewError(ErrUnknownResource, fmt.Sprintf("unknown resource %q", alias), nil)
}

func HTTPStatus(err error) int {
	var serviceErr *Error
	if !errors.As(err, &serviceErr) {
		return http.StatusInternalServerError
	}
	switch serviceErr.Kind {
	case ErrUnknownResource:
		return http.StatusNotFound
	case ErrWrongScope, ErrBadRequest:
		return http.StatusBadRequest
	case ErrCacheNotReady:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
```

- [ ] **Step 4: Implement service operations**

Create `api/resources/service/service.go` with `List`, `Get`, `Create`, `Patch`, and `Delete`. Use the registry to resolve aliases. For this task, implement `List`, `Get`, and `Create`; add method stubs for patch/delete that return `ErrBadRequest` until Task 5 handler tests drive exact behavior. `Create` must set namespace from the path for namespaced resources before calling `client.Create`.

- [ ] **Step 5: Run service tests and verify they pass**

Run: `go test ./api/resources/service -run Test -v`

Expected: PASS.

- [ ] **Step 6: Commit service**

Run:

```bash
git add api/resources/service
git commit -m "feat: add api resource service"
```

---

### Task 4: Watch Manager

**Files:**
- Create: `api/resources/watch/watcher.go`
- Test: `api/resources/watch/watcher_test.go`

- [ ] **Step 1: Write failing watcher tests**

Create `api/resources/watch/watcher_test.go` with a fake client that has initial configmaps. Verify one call to `SyncOnce` replaces the store scope and marks readiness.

- [ ] **Step 2: Run watcher tests and verify they fail**

Run: `go test ./api/resources/watch -run Test -v`

Expected: FAIL because watcher is not implemented.

- [ ] **Step 3: Implement `SyncOnce` before continuous watch**

Create `api/resources/watch/watcher.go` with:

- `type Manager struct { registry *registry.Registry; store *store.Store; client client.WithWatch; logger *zap.Logger }`
- `func NewManager(...) *Manager`
- `func (m *Manager) SyncOnce(ctx context.Context, resource registry.Resource, namespace string) error`
- `func (m *Manager) StartClusterScoped(ctx context.Context)` that runs `SyncOnce` for all cluster-scoped resources and logs errors.
- `func (m *Manager) EnsureNamespaced(ctx context.Context, alias string, namespace string) error` that runs `SyncOnce` for the requested namespace if not ready.

Implementation details:

- Use `resource.NewList()` and `client.List(ctx, list, client.InNamespace(namespace))` for namespaced resources.
- Convert list items using `meta.ExtractList` from `k8s.io/apimachinery/pkg/api/meta`.
- Convert each runtime object to `client.Object` and pass to `store.ReplaceScope`.

- [ ] **Step 4: Run watcher tests and verify they pass**

Run: `go test ./api/resources/watch -run Test -v`

Expected: PASS.

- [ ] **Step 5: Commit watcher sync**

Run:

```bash
git add api/resources/watch
git commit -m "feat: add api resource watch sync"
```

---

### Task 5: HTTP Handlers and Server Wiring

**Files:**
- Create: `api/resources/handlers/handlers.go`
- Test: `api/resources/handlers/handlers_test.go`
- Modify: `api/server/webhook_routes.go`
- Modify: `api/server/server.go`

- [ ] **Step 1: Write failing handler tests**

Create `api/resources/handlers/handlers_test.go` with Gin tests for:

- `GET /api/v1/namespaces/default/resources/configmaps` returns `503` before cache readiness.
- `GET /api/v1/resources/workspaces` returns `400` because workspaces are namespaced.
- `GET /api/v1/resources/does-not-exist` returns `404`.
- `POST /api/v1/namespaces/default/resources/configmaps` creates a configmap through the fake client and returns `201`.

- [ ] **Step 2: Run handler tests and verify they fail**

Run: `go test ./api/resources/handlers -run Test -v`

Expected: FAIL because handlers are not implemented.

- [ ] **Step 3: Implement handlers**

Create `api/resources/handlers/handlers.go` with:

- `type Handler struct { service *service.Service }`
- `func Register(router gin.IRoutes, svc *service.Service)` or `func (h *Handler) Register(group *gin.RouterGroup)`.
- Cluster routes and namespaced routes from the spec.
- JSON error response: `gin.H{"error": err.Error()}` using `service.HTTPStatus(err)`.
- Decode create/patch bodies into `unstructured.Unstructured`; set namespace from path for namespaced creates.
- List response shape: `gin.H{"items": items}` for store-returned objects.

- [ ] **Step 4: Wire handlers into API server router**

Modify `api/server/webhook_routes.go`:

- Change setup function signature to accept a resource service.
- Under existing `v1 := router.Group("/api/v1")`, call resource handler registration.
- Update `/api/v1/info` message from `CRUD operations are handled by Next.js Server Actions` to `resource operations are served by kloudlite-api`.

- [ ] **Step 5: Wire resource construction into server**

Modify `api/server/server.go`:

- Construct `registry.Default()`, `store.New()`, `service.New(...)`, and `watch.NewManager(...)` in `New`.
- Add the watch manager field to `Server`.
- Pass service into `setupWebhookRouter`.
- In `Start`, start cluster-scoped sync/watch in a goroutine before logging API server startup.

- [ ] **Step 6: Run focused handler/server tests**

Run:

```bash
go test ./api/resources/handlers ./api/server -run Test -v
```

Expected: PASS.

- [ ] **Step 7: Commit handler wiring**

Run:

```bash
git add api/resources/handlers api/server
git commit -m "feat: expose api resource routes"
```

---

### Task 6: Continuous Watch and Final Verification

**Files:**
- Modify: `api/resources/watch/watcher.go`
- Test: `api/resources/watch/watcher_test.go`
- Modify as needed: `api/resources/service/service.go`
- Modify as needed: `api/resources/handlers/handlers.go`

- [ ] **Step 1: Add failing watch event test**

Extend `api/resources/watch/watcher_test.go` with a fake watch test that sends `Added`, `Modified`, and `Deleted` events and verifies the store changes accordingly.

- [ ] **Step 2: Run watch tests and verify they fail**

Run: `go test ./api/resources/watch -run Test -v`

Expected: FAIL because continuous watch is not implemented.

- [ ] **Step 3: Implement continuous watch loop**

In `api/resources/watch/watcher.go`, add:

- `func (m *Manager) RunScope(ctx context.Context, resource registry.Resource, namespace string)`.
- Initial `SyncOnce`.
- `client.Watch(ctx, resource.NewList(), opts...)`.
- Event handling for `watch.Added`, `watch.Modified`, and `watch.Deleted`.
- On watch close/error, log and relist before retrying until context cancellation.

- [ ] **Step 4: Complete patch/delete service methods**

Drive with service tests, then implement:

- `Patch(ctx, alias, namespace, name, patchObject)` using `client.MergeFrom(current)` or raw merge patch if handlers pass patch bytes.
- `Delete(ctx, alias, namespace, name)` by constructing the registered object type with name/namespace and calling `client.Delete`.
- Map Kubernetes `IsNotFound`, `IsConflict`, and invalid/admission errors into typed service errors and HTTP statuses.

- [ ] **Step 5: Run focused resource tests**

Run:

```bash
go test ./api/resources/... -v
```

Expected: PASS.

- [ ] **Step 6: Run broad Go verification**

Run:

```bash
go test ./...
go list ./...
CGO_ENABLED=0 go build -ldflags='-s -w' -o /tmp/opencode/kloudlite ./cli/kloudlite
```

Expected: all commands PASS.

- [ ] **Step 7: Commit final resource layer**

Run:

```bash
git add api/resources api/server
git commit -m "feat: watch api resources in server"
```

---

## Self-Review Notes

- Spec coverage: registry, cache/store, API surface, error mapping, watch startup, and verification are covered by Tasks 1-6.
- Dashboard migration is intentionally excluded.
- Installer/deployment manifest changes are intentionally excluded.
- Runtime impact: `kloudlite server api` will open Kubernetes watches and serve resource endpoints from the same HTTPS server.

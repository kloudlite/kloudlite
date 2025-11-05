# Remaining Cluster-Scoped Migration Tasks

This document outlines the remaining tasks to complete the cluster-scoped resource migration for Workspace, Environment, ServiceIntercept, and DomainRequest resources.

## ✅ Completed (Core Controllers)

All core controller logic has been successfully migrated to cluster-scoped resources:

1. **Type Definitions** - All CRDs converted to cluster-scoped
2. **Controllers** - Workspace, Environment, ServiceIntercept, DomainRequest, WorkMachine controllers updated
3. **Ownership** - Proper OwnerReference hierarchies established
4. **Webhooks** - Validation and mutation webhooks updated
5. **Infrastructure** - kloudlite-ingress namespace created
6. **RBAC** - ClusterRole already has correct permissions

## 📋 Remaining Tasks (API Layer)

The following API-layer components still reference the old namespace-based approach:

### 1. Repository Layer

**Files to Update:**
- `/Users/karthik/dev/kl-workspace/kloudlite-v2/api/internal/repository/workspace_repository.go`
- `/Users/karthik/dev/kl-workspace/kloudlite-v2/api/internal/repository/environment_repository.go`
- `/Users/karthik/dev/kl-workspace/kloudlite-v2/api/internal/repository/k8s_repository.go` (ServiceIntercept methods)

**Changes Needed:**

```go
// OLD: WorkspaceRepository (namespace-scoped)
type WorkspaceRepository interface {
    NamespacedRepository[*workspacesv1.Workspace, *workspacesv1.WorkspaceList]
    GetByOwner(ctx context.Context, owner string, namespace string) (*workspacesv1.WorkspaceList, error)
    // ... more methods with namespace parameter
}

// NEW: WorkspaceRepository (cluster-scoped)
type WorkspaceRepository interface {
    ClusterRepository[*workspacesv1.Workspace, *workspacesv1.WorkspaceList]
    GetByOwner(ctx context.Context, owner string) (*workspacesv1.WorkspaceList, error)
    // ... methods without namespace parameter
}
```

**Implementation Updates:**
```go
// OLD
baseRepo := NewK8sNamespacedRepository(
    k8sClient,
    func() *workspacesv1.Workspace { return &workspacesv1.Workspace{} },
    func() *workspacesv1.WorkspaceList { return &workspacesv1.WorkspaceList{} },
)

// NEW
baseRepo := NewK8sClusterRepository(
    k8sClient,
    func() *workspacesv1.Workspace { return &workspacesv1.Workspace{} },
    func() *workspacesv1.WorkspaceList { return &workspacesv1.WorkspaceList{} },
)
```

### 2. API Handlers

**Files to Update:**
- `/Users/karthik/dev/kl-workspace/kloudlite-v2/api/internal/handlers/workspace_handlers.go`
- `/Users/karthik/dev/kl-workspace/kloudlite-v2/api/internal/handlers/environment_handlers.go`
- `/Users/karthik/dev/kl-workspace/kloudlite-v2/api/internal/handlers/serviceintercept_handlers.go`

**Changes Needed:**

```go
// OLD: API routes with namespace parameter
// GET /api/v1/namespaces/:namespace/workspaces/:name
func (h *WorkspaceHandlers) GetWorkspace(c *gin.Context) {
    namespace := c.Param("namespace")
    name := c.Param("name")
    workspace, err := h.wsRepo.Get(c.Request.Context(), namespace, name)
    // ...
}

// NEW: API routes without namespace parameter
// GET /api/v1/workspaces/:name
func (h *WorkspaceHandlers) GetWorkspace(c *gin.Context) {
    name := c.Param("name")
    workspace, err := h.wsRepo.Get(c.Request.Context(), name)
    // ...
}
```

**Route Updates:**
```go
// OLD routes
v1.GET("/namespaces/:namespace/workspaces", workspaceHandlers.ListWorkspaces)
v1.GET("/namespaces/:namespace/workspaces/:name", workspaceHandlers.GetWorkspace)
v1.POST("/namespaces/:namespace/workspaces", workspaceHandlers.CreateWorkspace)

// NEW routes
v1.GET("/workspaces", workspaceHandlers.ListWorkspaces)
v1.GET("/workspaces/:name", workspaceHandlers.GetWorkspace)
v1.POST("/workspaces", workspaceHandlers.CreateWorkspace)
```

### 3. Domain-Specific Methods

Methods like `GetByOwner`, `GetByWorkMachine` need to filter cluster-wide results:

```go
// Example: GetByOwner
func (r *workspaceRepository) GetByOwner(ctx context.Context, owner string) (*workspacesv1.WorkspaceList, error) {
    return r.List(ctx, WithFieldSelector("spec.owner="+owner))
}

// Example: GetByWorkMachine (now actually useful since cluster-scoped)
func (r *workspaceRepository) GetByWorkMachine(ctx context.Context, workMachineName string) (*workspacesv1.WorkspaceList, error) {
    return r.List(ctx, WithFieldSelector("spec.workmachineName="+workMachineName))
}
```

## 🧪 Testing Checklist

After completing the above changes, test the following:

### Controller Tests
- [ ] Workspace creation with WorkMachine ownership
- [ ] Environment creation with WorkMachine ownership
- [ ] ServiceIntercept creation (cluster-scoped)
- [ ] DomainRequest with kloudlite-ingress workloads
- [ ] Cascading deletion when WorkMachine is deleted

### API Tests
- [ ] Create workspace via API (no namespace in URL)
- [ ] List workspaces cluster-wide
- [ ] Filter workspaces by owner
- [ ] Filter workspaces by WorkMachine
- [ ] Update workspace status
- [ ] Delete workspace

### Integration Tests
- [ ] Full workspace lifecycle (create → activate → suspend → delete)
- [ ] Environment lifecycle with namespace management
- [ ] Service intercept setup and teardown
- [ ] Multi-user isolation (workspaces in different WorkMachines)

## 📝 Migration Guide

For existing deployments, the migration would involve:

1. **Backup existing resources**
2. **Apply new CRDs** (cluster-scoped)
3. **Migrate data** (copy resources to cluster scope)
4. **Update ownership** (add WorkMachine OwnerReferences)
5. **Test controllers** (verify reconciliation)
6. **Deploy new API** (with updated handlers)
7. **Verify** (end-to-end testing)

## 🔧 Quick Implementation Script

```bash
# Update WorkspaceRepository
sed -i 's/NamespacedRepository/ClusterRepository/g' internal/repository/workspace_repository.go
sed -i 's/, namespace string//g' internal/repository/workspace_repository.go
sed -i 's/NewK8sNamespacedRepository/NewK8sClusterRepository/g' internal/repository/workspace_repository.go

# Similar updates for Environment and ServiceIntercept repositories

# Update handlers
find internal/handlers -name "*_handlers.go" -exec sed -i 's/c.Param("namespace")//g' {} \;

# Update routes
# Manual editing required for route definitions
```

## 📚 References

- Kubernetes API Conventions: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md
- Controller Runtime: https://pkg.go.dev/sigs.k8s.io/controller-runtime
- Kubebuilder Markers: https://book.kubebuilder.io/reference/markers.html

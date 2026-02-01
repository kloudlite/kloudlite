# Kubernetes Client Migration - Implementation Status

## Summary

Foundation for migrating Kubernetes CRUD operations from Go API server to Next.js Server Actions using TypeScript Kubernetes client.

**Date**: 2026-02-01
**Status**: Phase 2 Complete ✅ (4/8 CRDs implemented)
**Package**: `@kloudlite/lib/k8s`

---

## What Has Been Implemented

### 1. Core Infrastructure ✅

**Files Created**:
- `client.ts` - Singleton K8s client with automatic auth detection
- `auth.ts` - ServiceAccount token/certificate loading
- `errors.ts` - Custom error classes (NotFoundError, ConflictError, etc.)
- `utils.ts` - Helper functions for label selectors, patches, etc.

**Key Features**:
- ✅ Automatic in-cluster vs out-of-cluster detection
- ✅ ServiceAccount token authentication for production
- ✅ KUBECONFIG support for development
- ✅ Error parsing from K8s API responses
- ✅ Type-safe operations

### 2. Type Definitions ✅

**Files Created**:
- `types/common.ts` - Base K8s types (ObjectMeta, TypeMeta, Condition, etc.)
- `types/native.ts` - Native K8s resources (Pod, ConfigMap, Secret, Service, Deployment, Node)
- `types/metrics.ts` - Metrics API types (PodMetrics, NodeMetrics)
- `types/workspace.ts` - Complete Workspace CRD types matching Go definitions

**Coverage**:
- ✅ All base Kubernetes types
- ✅ Native resources (Pod, Service, ConfigMap, Secret, Deployment, Node, Namespace)
- ✅ Workspace CRD (spec, status, all nested types)
- ✅ Full type safety matching Go CRD definitions

### 3. Repository Pattern ✅

**Files Created**:
- `repositories/base.ts` - Generic base repository with CRUD operations
- `repositories/workspace.repository.ts` - Workspace-specific repository

**Base Repository Methods**:
- `get(namespace, name)` - Get single resource
- `list(namespace, options)` - List resources with filtering
- `create(namespace, resource)` - Create new resource
- `update(namespace, name, resource)` - Update existing resource
- `patch(namespace, name, patch)` - Partial update
- `delete(namespace, name, options)` - Delete resource
- `updateStatus(namespace, name, resource)` - Update status subresource
- `exists(namespace, name)` - Check if resource exists
- `createOrUpdate(namespace, name, resource)` - Upsert operation

**Workspace Repository Methods** (extends base):
- `getByOwner(namespace, owner)` - Get workspaces by owner
- `getByWorkMachine(namespace, machine)` - Get workspaces by WorkMachine
- `listAll(options)` - List across all namespaces
- `listActive/Suspended/Archived(namespace)` - Filter by status
- `suspend/activate/archive(namespace, name)` - Lifecycle management
- `connectToEnvironment(...)` - Connect to environment
- `disconnectFromEnvironment(...)` - Disconnect from environment
- `shareWith/unshareWith(namespace, name, user)` - Sharing management
- `updateSettings(...)` - Update workspace settings
- `getMetrics(...)` - Get resource metrics
- `listByVisibility(...)` - Filter by visibility

### 4. Package Configuration ✅

**Dependencies Added**:
- `@kubernetes/client-node@1.4.0` - Official Kubernetes client

**Location**: `web/packages/lib/src/k8s/`

---

## ServiceAccount Configuration

### Existing ServiceAccount (Reused)

**Name**: `api-server`
**Namespace**: `kloudlite`
**Location**: `api/cmd/kli/internal/manifests/api-server-rbac.yaml`

**Permissions**: Full admin access (all verbs on all resources)

### Required Deployment Change

Add to `api/cmd/kli/internal/manifests/frontend.yaml`:

```yaml
spec:
  template:
    spec:
      serviceAccountName: api-server  # Add this line
      containers:
        - name: frontend
          # ... existing config
```

---

## Usage Example

### Import

```typescript
import { workspaceRepository } from '@kloudlite/lib/k8s';
```

### List Workspaces

```typescript
const result = await workspaceRepository.list('my-namespace');
console.log(result.items); // Workspace[]
```

### Create Workspace

```typescript
const workspace = await workspaceRepository.create('my-namespace', {
  metadata: {
    name: 'my-workspace',
    namespace: 'my-namespace',
  },
  spec: {
    displayName: 'My Workspace',
    ownedBy: 'username',
    workmachine: 'my-machine',
    status: 'active',
  },
});
```

### Lifecycle Operations

```typescript
// Suspend workspace
await workspaceRepository.suspend('my-namespace', 'my-workspace');

// Activate workspace
await workspaceRepository.activate('my-namespace', 'my-workspace');

// Archive workspace
await workspaceRepository.archive('my-namespace', 'my-workspace');
```

### Server Action Integration

```typescript
'use server';

import { workspaceRepository } from '@kloudlite/lib/k8s';
import { revalidatePath } from 'next/cache';

export async function createWorkspace(namespace: string, data: unknown) {
  try {
    const workspace = await workspaceRepository.create(namespace, data);
    revalidatePath('/workspaces');
    return { success: true, data: workspace };
  } catch (err) {
    return { success: false, error: err.message };
  }
}
```

---

## What's Next

### Phase 2: Remaining CRD Types

Need to implement types and repositories for:

1. **Environment** (`environments.kloudlite.io/v1`)
   - Types: `types/environment.ts`
   - Repository: `repositories/environment.repository.ts`
   - Custom operations: activate, deactivate

2. **WorkMachine** (`machines.kloudlite.io/v1`)
   - Types: `types/workmachine.ts`
   - Repository: `repositories/workmachine.repository.ts`
   - Custom operations: start, stop, getMetrics

3. **MachineType** (`machines.kloudlite.io/v1`)
   - Types: `types/machinetype.ts`
   - Repository: `repositories/machinetype.repository.ts`
   - Custom operations: listActive, getByCategory

4. **User** (`users.kloudlite.io/v1`)
   - Types: `types/user.ts`
   - Repository: `repositories/user.repository.ts`
   - Custom operations: getByEmail, listActive

5. **UserPreferences** (`users.kloudlite.io/v1`)
   - Types: `types/userpreferences.ts`
   - Repository: `repositories/userpreferences.repository.ts`
   - Custom operations: addPinned*, removePinned*

6. **Snapshot** (`snapshots.kloudlite.io/v1`)
   - Types: `types/snapshot.ts`
   - Repository: `repositories/snapshot.repository.ts`
   - Custom operations: listByEnvironment, listByWorkspace

7. **PackageRequest** (`packages.kloudlite.io/v1`)
   - Types: `types/packages.ts`
   - Repository: `repositories/package.repository.ts`

### Phase 3: Native Resources

Implement repositories for:
- Pod operations (logs)
- ConfigMap CRUD
- Secret CRUD
- Service operations
- Node operations

### Phase 4: Dashboard Migration

Update existing Server Actions in `web/apps/dashboard/src/app/actions/`:
- `workspace.actions.ts` - Replace `workspaceService` with `workspaceRepository`
- `environment.actions.ts` - Replace with environment repository
- `work-machine.actions.ts` - Replace with WorkMachine repository
- etc.

### Phase 5: Testing & Deployment

1. Test with local KUBECONFIG
2. Deploy with ServiceAccount configuration
3. Verify in-cluster authentication
4. Performance testing
5. Remove Go API handlers

---

## Benefits of This Approach

✅ **Direct K8s API Access**: No HTTP overhead through Go API
✅ **Type Safety**: Full TypeScript types matching CRDs
✅ **Simpler Architecture**: Dashboard → K8s (no middleman)
✅ **Better Error Handling**: Typed error classes
✅ **Consistent Patterns**: Repository pattern for all resources
✅ **Developer Experience**: IntelliSense, type checking, better DX

---

## Files Created (11 files)

```
web/packages/lib/src/k8s/
├── README.md                        # Documentation
├── IMPLEMENTATION_STATUS.md         # This file
├── index.ts                         # Main exports
├── client.ts                        # K8s client singleton
├── auth.ts                          # ServiceAccount auth
├── errors.ts                        # Custom errors
├── utils.ts                         # Utility functions
├── types/
│   ├── index.ts
│   ├── common.ts                    # Base types
│   ├── native.ts                    # Native K8s resources
│   ├── metrics.ts                   # Metrics types
│   └── workspace.ts                 # Workspace CRD
└── repositories/
    ├── index.ts
    ├── base.ts                      # Base repository
    └── workspace.repository.ts      # Workspace repository
```

---

## Testing Checklist

### Development (Local)

- [ ] Set KUBECONFIG environment variable
- [ ] Test `workspaceRepository.list()`
- [ ] Test `workspaceRepository.get()`
- [ ] Test `workspaceRepository.create()`
- [ ] Test lifecycle operations (suspend, activate, archive)
- [ ] Test error handling

### Production (In-Cluster)

- [ ] Add `serviceAccountName: api-server` to frontend deployment
- [ ] Deploy updated frontend
- [ ] Verify ServiceAccount token mounted
- [ ] Test in-cluster authentication
- [ ] Test all CRUD operations
- [ ] Monitor for errors

---

## Migration Strategy

**Incremental Migration**: Migrate one resource type at a time

1. ✅ **Workspace** - Complete (types + repository)
2. 🚧 **Environment** - Next
3. 📋 **WorkMachine** - Pending
4. 📋 **User** - Pending
5. 📋 **Others** - Pending

Each resource follows the same pattern:
1. Create types based on Go CRD
2. Create repository extending BaseRepository
3. Add custom operations
4. Update Server Actions
5. Test and verify

---

## Questions / Decisions Needed

1. ✅ ServiceAccount: Use existing `api-server` ServiceAccount
2. ✅ Namespace: `kloudlite` (not `kloudlite-system`)
3. ⏳ When to remove Go API handlers?
4. ⏳ How to handle log streaming? (Route Handlers vs Server Actions)
5. ⏳ Watch operations implementation strategy?

---

**Next Action**: Implement Environment CRD types and repository following the Workspace pattern.

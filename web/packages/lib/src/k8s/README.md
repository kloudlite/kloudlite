# Kubernetes Client Library for Kloudlite

TypeScript client library for interacting with Kubernetes API, including Kloudlite custom resources.

## Overview

This library provides a complete TypeScript interface to Kubernetes, designed to replace the Go API server for CRUD operations. It uses the official `@kubernetes/client-node` library with ServiceAccount authentication for in-cluster access.

## Features

- ✅ **ServiceAccount Authentication**: In-cluster authentication using mounted ServiceAccount tokens
- ✅ **Custom Resource Support**: Full TypeScript types for Kloudlite CRDs (Workspace, Environment, etc.)
- ✅ **Repository Pattern**: Clean abstraction over Kubernetes API with domain-specific methods
- ✅ **Error Handling**: Custom error classes with proper HTTP status codes
- ✅ **Type Safety**: Complete TypeScript type definitions matching Go CRD structures

## Architecture

```
web/packages/lib/src/k8s/
├── client.ts                   # Singleton K8s client with auth
├── auth.ts                     # ServiceAccount token/cert loader
├── errors.ts                   # Custom error classes
├── utils.ts                    # Helper functions
├── types/                      # TypeScript type definitions
│   ├── common.ts              # Base K8s types (ObjectMeta, etc.)
│   ├── native.ts              # Native K8s resources (Pod, Service, etc.)
│   ├── metrics.ts             # Metrics API types
│   └── workspace.ts           # Workspace CRD types ✅
└── repositories/               # Repository pattern implementations
    ├── base.ts                # Generic CRUD base class ✅
    └── workspace.repository.ts # Workspace repository ✅
```

## Installation

The library is already part of `@kloudlite/lib` package:

```bash
# In web/ directory
bun install  # Dependencies already added
```

## Usage

### Basic Usage

```typescript
import { getK8sClient, workspaceRepository } from '@kloudlite/lib/k8s';

// Get singleton client
const client = getK8sClient();

// Use repositories
const workspaces = await workspaceRepository.list('my-namespace');
const workspace = await workspaceRepository.get('my-namespace', 'my-workspace');
```

### Server Actions (Next.js)

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

### Repository Methods

All repositories extend `BaseRepository` with standard CRUD operations:

```typescript
// CRUD operations
await repository.get(namespace, name);
await repository.list(namespace, options);
await repository.create(namespace, resource);
await repository.update(namespace, name, resource);
await repository.patch(namespace, name, patch);
await repository.delete(namespace, name, options);
await repository.updateStatus(namespace, name, resource);

// Helper methods
await repository.exists(namespace, name);
await repository.createOrUpdate(namespace, name, resource);
```

### Workspace-Specific Operations

```typescript
// Lifecycle management
await workspaceRepository.suspend(namespace, name);
await workspaceRepository.activate(namespace, name);
await workspaceRepository.archive(namespace, name);

// Environment connection
await workspaceRepository.connectToEnvironment(namespace, name, envName);
await workspaceRepository.disconnectFromEnvironment(namespace, name);

// Sharing
await workspaceRepository.shareWith(namespace, name, username);
await workspaceRepository.unshareWith(namespace, name, username);

// Filtering
await workspaceRepository.listActive(namespace);
await workspaceRepository.listSuspended(namespace);
await workspaceRepository.getByOwner(namespace, owner);
```

## Implementation Status

### ✅ Completed (Phase 1)

- [x] K8s client with ServiceAccount auth (`client.ts`)
- [x] Authentication helpers (`auth.ts`)
- [x] Error handling (`errors.ts`)
- [x] Utility functions (`utils.ts`)
- [x] Base repository pattern (`repositories/base.ts`)
- [x] Common types (`types/common.ts`)
- [x] Native K8s types (`types/native.ts`)
- [x] Metrics types (`types/metrics.ts`)
- [x] Workspace CRD types (`types/workspace.ts`)
- [x] Workspace repository (`repositories/workspace.repository.ts`)
- [x] Verified existing `api-server` ServiceAccount for reuse

### 🚧 In Progress (Phase 2)

- [ ] Environment CRD types (`types/environment.ts`)
- [ ] Environment repository (`repositories/environment.repository.ts`)
- [ ] WorkMachine CRD types (`types/workmachine.ts`)
- [ ] WorkMachine repository (`repositories/workmachine.repository.ts`)
- [ ] MachineType CRD types (`types/machinetype.ts`)
- [ ] MachineType repository (`repositories/machinetype.repository.ts`)
- [ ] User CRD types (`types/user.ts`)
- [ ] User repository (`repositories/user.repository.ts`)
- [ ] UserPreferences CRD types (`types/userpreferences.ts`)
- [ ] UserPreferences repository (`repositories/userpreferences.repository.ts`)
- [ ] Snapshot CRD types (`types/snapshot.ts`)
- [ ] Snapshot repository (`repositories/snapshot.repository.ts`)

### 📋 Pending (Phase 3+)

- [ ] Native resource repositories (Pod, ConfigMap, Secret, etc.)
- [ ] Update dashboard Server Actions to use repositories
- [ ] Log streaming via Route Handlers
- [ ] Watch operations for real-time updates
- [ ] Metrics streaming
- [ ] Delete Go API handlers and repositories
- [ ] Update deployment manifests
- [ ] Production testing

## Configuration

### Development (Out-of-Cluster)

Set `KUBECONFIG` environment variable:

```bash
export KUBECONFIG=/path/to/kubeconfig
# Or use default: ~/.kube/config
```

### Production (In-Cluster)

The client automatically detects in-cluster environment and uses the existing `api-server` ServiceAccount:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend
  namespace: kloudlite
spec:
  template:
    spec:
      serviceAccountName: api-server  # Reuses existing api-server ServiceAccount
      containers:
        - name: frontend
          # ServiceAccount token auto-mounted at:
          # /var/run/secrets/kubernetes.io/serviceaccount/token
```

**Note**: To enable K8s client in the frontend deployment, add `serviceAccountName: api-server` to the frontend deployment spec in `api/cmd/kli/internal/manifests/frontend.yaml`.

## RBAC Permissions

The dashboard reuses the existing `api-server` ServiceAccount in the `kloudlite` namespace, which already has full admin permissions to all Kubernetes resources.

**ServiceAccount**: `api-server` (namespace: `kloudlite`)
**Location**: `api/cmd/kli/internal/manifests/api-server-rbac.yaml`

No additional RBAC configuration needed - the existing ServiceAccount has all required permissions.

## Error Handling

The library provides custom error classes:

```typescript
import { NotFoundError, ConflictError, parseK8sError } from '@kloudlite/lib/k8s';

try {
  await workspaceRepository.get(namespace, name);
} catch (err) {
  if (err instanceof NotFoundError) {
    console.log('Workspace not found');
  } else if (err instanceof ConflictError) {
    console.log('Workspace already exists');
  } else {
    console.error('Unknown error:', err);
  }
}
```

## Type Safety

All resources are fully typed:

```typescript
import type { Workspace, WorkspaceSpec, WorkspaceStatus } from '@kloudlite/lib/k8s';

const workspace: Workspace = {
  apiVersion: 'workspaces.kloudlite.io/v1',
  kind: 'Workspace',
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
};
```

## Next Steps

1. **Add ServiceAccount to frontend deployment**: Add `serviceAccountName: api-server` to `frontend.yaml`
2. **Implement remaining CRD types**: Environment, WorkMachine, User, etc.
3. **Create remaining repositories**: Following the same pattern as WorkspaceRepository
4. **Update Server Actions**: Replace API calls with repository calls
5. **Test in development**: Ensure KUBECONFIG works locally
6. **Production testing**: Verify in-cluster authentication works
7. **Cleanup Go code**: Remove API handlers and repositories

## Migration Guide

### Before (Go API)

```typescript
// Using Go API via HTTP
const response = await fetch(`${apiUrl}/api/v1/namespaces/${namespace}/workspaces`, {
  headers: { Authorization: `Bearer ${token}` },
});
const workspaces = await response.json();
```

### After (K8s Client)

```typescript
// Using K8s client directly
const result = await workspaceRepository.list(namespace);
const workspaces = result.items;
```

### Benefits

- ✅ No HTTP overhead (direct K8s API calls)
- ✅ Type-safe operations
- ✅ Automatic authentication (ServiceAccount)
- ✅ Better error handling
- ✅ Simpler architecture (no Go API middleman)

## Contributing

When adding new CRD types:

1. Create type definition in `types/{resource}.ts`
2. Create repository in `repositories/{resource}.repository.ts`
3. Export from `types/index.ts` and `repositories/index.ts`
4. Update RBAC manifest if needed
5. Create corresponding Server Actions in dashboard

## License

Part of the Kloudlite project.

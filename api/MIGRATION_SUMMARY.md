# API Server Migration Summary

## Overview

The API server has been cleaned up to remove all HTTP CRUD handlers. All CRUD operations are now handled by Next.js Server Actions using the TypeScript Kubernetes client.

**Date**: February 1, 2026
**Migration Type**: HTTP API → Next.js Server Actions

---

## What Remains in API Server

### 1. **Controllers** (Reconciliation Loops)
All Kubernetes controllers remain unchanged and continue to run:
- User controller
- Environment controller
- WorkMachine controller
- Workspace controller
- Snapshot controller
- EnvironmentSnapshotRequest controller
- EnvironmentSnapshotRestore controller
- EnvironmentForkRequest controller
- PackageRequest controller

**Location**: `api/internal/controllers/`

### 2. **Webhooks** (Kubernetes Admission Webhooks)
All validation and mutation webhooks remain active on HTTPS port 8443:
- User webhooks (validate/mutate)
- Environment webhooks (validate/mutate)
- MachineType webhooks (validate/mutate)
- WorkMachine webhooks (validate/mutate)
- Workspace webhooks (validate/mutate)
- ConfigMap webhooks (validate)
- Secret webhooks (validate)
- Service webhooks (mutate)
- Pod webhooks (mutate)
- Snapshot webhooks (validate)

**Location**: `api/internal/webhooks/`

### 3. **Health Checks**
- `/health` - Health check endpoint
- `/ready` - Readiness check endpoint

### 4. **Info Endpoint**
- `/api/v1/info` - Returns server mode and status

---

## What Was Removed

### HTTP API Routes (All CRUD Operations)
The following routes have been removed as they're now handled by Next.js Server Actions:

#### User Management
- `POST /api/v1/users`
- `GET /api/v1/users/:name`
- `PUT /api/v1/users/:name`
- `DELETE /api/v1/users/:name`
- `POST /api/v1/users/:name/reset-password`
- `POST /api/v1/users/:name/activate`
- `POST /api/v1/users/:name/deactivate`
- `GET /api/v1/users`

#### Environment Management
- `POST /api/v1/environments`
- `GET /api/v1/environments/:name`
- `PUT /api/v1/environments/:name`
- `DELETE /api/v1/environments/:name`
- `POST /api/v1/environments/:name/activate`
- `POST /api/v1/environments/:name/deactivate`
- All environment config/secret/file routes
- All environment snapshot routes

#### Workspace Management
- `POST /api/v1/namespaces/:namespace/workspaces`
- `GET /api/v1/namespaces/:namespace/workspaces/:name`
- `PUT /api/v1/namespaces/:namespace/workspaces/:name`
- `DELETE /api/v1/namespaces/:namespace/workspaces/:name`
- `POST /api/v1/namespaces/:namespace/workspaces/:name/suspend`
- `POST /api/v1/namespaces/:namespace/workspaces/:name/activate`
- All workspace package/snapshot routes

#### WorkMachine Management
- `GET /api/v1/work-machines/my`
- `POST /api/v1/work-machines/my`
- `PUT /api/v1/work-machines/my`
- `DELETE /api/v1/work-machines/my`
- `POST /api/v1/work-machines/my/start`
- `POST /api/v1/work-machines/my/stop`

#### MachineType Management
- `GET /api/v1/machine-types`
- `POST /api/v1/machine-types`
- `PUT /api/v1/machine-types/:name`
- `DELETE /api/v1/machine-types/:name`

#### Other Routes
- Authentication routes (`/api/v1/auth/*`)
- Dashboard routes (`/api/v1/dashboard`)
- User preferences routes
- Service routes
- Registry catalog routes
- VPN routes (temporarily disabled)

### Archived Files

#### Handler Files (Moved to `api/internal/handlers/archived/`)
- `auth_handlers.go` + test
- `dashboard_handlers.go`
- `environment_config_handlers.go` + test
- `environment_handlers.go` + test
- `machinetype_handlers.go` + test
- `oauth_handlers.go` + test
- `registry_catalog_handlers.go`
- `service_handlers.go` + test
- `snapshot_handlers.go`
- `superadmin_login_handlers.go`
- `user_handlers.go` + test
- `userpreferences_handlers.go`
- `vpn_handlers.go`
- `workmachine_handlers.go` + test
- `workspace_handlers.go` + test

#### Server Files (Moved to `api/internal/server/`)
- `routes.go.archived` - Old comprehensive HTTP routes

### Removed Dependencies
- Repository Manager - No longer needed (controllers have their own K8s clients)
- Services Manager - Only used by HTTP handlers
- HTTP server on port 8080 - All CRUD now in Next.js

---

## Architecture Changes

### Before
```
┌─────────────────────────────────────────┐
│         Go API Server                   │
│  ┌───────────────────────────────────┐  │
│  │   HTTP Server :8080               │  │
│  │   - CRUD API Routes               │  │
│  │   - User/Env/WM/Workspace handlers│  │
│  └───────────────────────────────────┘  │
│  ┌───────────────────────────────────┐  │
│  │   HTTPS Server :8443              │  │
│  │   - Webhooks                      │  │
│  │   - All HTTP routes               │  │
│  └───────────────────────────────────┘  │
│  ┌───────────────────────────────────┐  │
│  │   Controller Manager              │  │
│  │   - Reconciliation loops          │  │
│  └───────────────────────────────────┘  │
│  ┌───────────────────────────────────┐  │
│  │   Repository Manager              │  │
│  │   - K8s CRUD operations           │  │
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘
```

### After
```
┌─────────────────────────────────────────┐
│    Next.js Dashboard (Node.js)          │
│  ┌───────────────────────────────────┐  │
│  │   Server Actions                  │  │
│  │   - Direct K8s client             │  │
│  │   - All CRUD operations           │  │
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘
              ↓ (ServiceAccount)
        Kubernetes API
              ↑
┌─────────────────────────────────────────┐
│         Go API Server                   │
│  ┌───────────────────────────────────┐  │
│  │   HTTPS Server :8443              │  │
│  │   - Webhooks ONLY                 │  │
│  │   - Health checks                 │  │
│  └───────────────────────────────────┘  │
│  ┌───────────────────────────────────┐  │
│  │   Controller Manager              │  │
│  │   - Reconciliation loops          │  │
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘
```

---

## Files Modified

### Core Server Files
1. **`api/internal/server/server.go`**
   - Removed HTTP server on port 8080
   - Removed repository manager
   - Removed services manager
   - Kept only HTTPS server for webhooks
   - Kept controller manager

2. **`api/internal/server/webhook_routes.go`** (New)
   - Minimal router with only webhooks
   - Health check endpoints
   - Info endpoint
   - No CRUD routes

3. **`api/cmd/server/main.go`**
   - No changes needed (server.New() and srv.Start() still work)

---

## Server Startup Process

### Before
1. Start HTTP server on :8080 (API routes)
2. Start HTTPS server on :8443 (API routes + webhooks)
3. Install webhook configurations
4. Start controller manager

### After
1. Start controller manager
2. Start HTTPS server on :8443 (webhooks only)
3. Install webhook configurations
4. Log: "API server started successfully (mode: controllers+webhooks)"

---

## Verification

### Build Test
```bash
cd api
go build ./cmd/server/main.go
# ✓ Compiles successfully (139MB binary)
```

### What Still Works
- ✅ All Kubernetes controllers reconcile resources
- ✅ Webhooks validate/mutate K8s resources
- ✅ Health checks respond
- ✅ Controller manager runs independently
- ✅ No dependencies on HTTP API handlers

### What's Moved to Next.js
- ✅ User authentication (via K8s User CRD)
- ✅ WorkMachine CRUD (via Server Actions)
- ✅ MachineType listing (via Server Actions)
- ✅ User preferences (via Server Actions)
- ✅ All dashboard data fetching

---

## Future Considerations

### VPN Endpoints
Currently disabled in `webhook_routes.go`:
```go
// VPN endpoints are currently disabled - uncomment if VPN service is re-enabled
```

If VPN functionality is needed, consider:
1. Creating a separate VPN service
2. Moving VPN to kltun CLI directly
3. Re-enabling VPN handlers with standalone service

### Metrics and Monitoring
- Controller manager has its own metrics (disabled to avoid port conflicts)
- Health checks available at `/health` and `/ready`
- Webhook server runs on :8443

### Deployment
- Ensure webhook certificates are valid
- Ensure ServiceAccount has correct RBAC permissions
- Ensure webhook server is accessible from K8s API server

---

## Benefits of Migration

1. **Simplified Architecture**: API server only handles what it needs (controllers + webhooks)
2. **Reduced Complexity**: No more HTTP API layer, repository manager, or services manager
3. **Better Separation**: CRUD operations in frontend, reconciliation in backend
4. **Easier Maintenance**: Each component has clear responsibility
5. **Performance**: Direct K8s client access from Next.js (no API server hop)
6. **Type Safety**: TypeScript types match K8s CRDs exactly

---

## Rollback Plan

If needed, archived files can be restored:
```bash
cd api/internal/handlers
mv archived/* .
rmdir archived

cd ../server
mv routes.go.archived routes.go
rm webhook_routes.go

# Restore server.go from git history
git checkout HEAD~1 api/internal/server/server.go
```

---

## Notes

- All handler tests are preserved in `archived/` directory
- VPN functionality temporarily disabled (can be re-enabled)
- Controllers remain completely unchanged
- Webhooks are critical infrastructure (must not be removed)

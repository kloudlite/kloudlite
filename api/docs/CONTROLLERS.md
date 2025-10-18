# Controllers

Controllers watch Kubernetes resources and reconcile actual state with desired state.

## User Controller

**File**: `internal/controllers/user/user_controller.go`

### What It Does
- Hashes passwords when `PasswordString` is set
- Updates user status (active/inactive)
- Manages finalizers for cleanup

### Password Flow
When `user.Spec.PasswordString` is set:
1. Hash password with SHA-256 for comparison
2. If changed, hash with bcrypt and update `user.Spec.Password`
3. Clear `PasswordString` field
4. Update status with hash for change detection

### Status Updates
- Sets `user.Status.Phase` to "active" or "inactive"
- Tracks creation time in `user.Status.CreatedAt`

**Note**: WorkMachine auto-creation is handled by UserService in handlers, not here.

---

## Environment Controller

**File**: `internal/controllers/environment/environment_controller.go`

### What It Does
- Creates/manages Kubernetes namespaces
- Handles activation/deactivation
- Automatically cleans up on deactivation
- Clones compositions from other environments

### Namespace Flow
1. Webhook sets `spec.targetNamespace` (e.g., `env-production`)
2. Webhook validates namespace doesn't already exist
3. Controller creates the actual Kubernetes namespace resource
4. Controller applies labels, annotations, resource quotas

### Activation/Deactivation
**Activate**: Scales all deployments from 0 to original replicas
**Deactivate**:
1. Disconnects all workspaces from this environment
2. Deletes all service intercepts in the namespace
3. Scales all deployments to 0 (handled by composition controller)

### Cleanup on Deactivation
Automatically runs before status update:
- Finds workspaces with `status.connectedEnvironment` matching this environment
- Clears their connection status
- Deletes all ServiceIntercepts in the environment namespace

---

## WorkMachine Controller

**File**: `internal/controllers/workmachine/workmachine_controller.go`

### What It Does
- Manages WorkMachine pod lifecycle based on `DesiredState`
- Updates status with pod details
- Propagates nodeSelector to host-manager deployment

### State Flow
- `DesiredState: stopped` → Delete pod
- `DesiredState: starting` → Create pod
- `DesiredState: running` → Ensure pod running
- `DesiredState: stopping` → Gracefully stop pod

### NodeSelector Propagation
Applies WorkMachine's `spec.nodeSelector` to host-manager deployment:
- Ensures host-manager runs on same node as other resources
- Enables shared Nix store access via hostPath volumes
- Only creates deployments, doesn't update existing ones

**Note**: If nodeSelector changes, delete deployment to pick up new value.

---

## Workspace Controller

**File**: `internal/controllers/workspace/workspace_controller.go`

### What It Does
- Manages workspace pod lifecycle
- Creates PackageRequest for Nix packages
- Handles auto-suspend
- Manages DNS for environment connections
- Creates services for workspace access

### Active Workspace Flow
1. Check if pod exists
2. Create pod with volumes (workspace data, SSH keys, Nix packages)
3. Create main service (ClusterIP)
4. Create headless service (for direct pod access)

### Package Management
Creates PackageRequest with:
- Profile name: `{workspace-name}-packages`
- Package list from `spec.packages`
- Watches PackageRequest status and updates workspace status

### Auto-Suspend
When `AutoSuspend` is configured:
- Monitors SSH, VS Code, TTYd connections
- Tracks idle time
- Deletes pod when `idleTime > spec.autoSuspend.idleTimeout`
- Updates status to "suspended"

### Cleanup on Deletion
**Files**: `internal/controllers/workspace/cleanup.go`

#### Host Directory Deletion
Uses cleanup pod pattern:
1. Validates path is safe (`validateHostPath`)
   - Must be within `/home/kl/workspaces/`
   - Must end with workspace name
   - Prevents path traversal (`..`)
2. Creates temporary pod with:
   - Alpine image with `rm -rf` command
   - HostPath volume mount to `/home/kl`
   - ActiveDeadlineSeconds (300s timeout)
   - Labels: `app=workspace-cleanup`, `type=temporary`
3. Pod deletes directory and self-terminates

#### Suspended Workspace Handling
- Deletes pod if exists
- Updates status to "Stopped"
- Clears pod-related status fields
- Requeues for verification

---

## Composition Controller

**File**: `internal/controllers/composition/composition_controller.go`

### What It Does
- Converts docker-compose to Kubernetes resources
- Creates deployments, services, ConfigMaps, secrets
- Injects environment variables
- Handles resource overrides
- Propagates nodeSelector from WorkMachine

### Conversion Flow
1. Parse `spec.composeContent` (YAML)
2. Convert each service to Deployment
3. Create Service for exposed ports
4. Inject `envFrom` (ConfigMaps/Secrets)
5. Apply `resourceOverrides` (replicas, resources)

### NodeSelector Propagation
1. Get environment for composition's namespace
2. Look up WorkMachine by environment creator
   - Sanitizes creator email: `john@example.com` → `wm-john-at-example-com`
   - Uses same logic as WorkMachine webhook
3. Fetch WorkMachine's `spec.nodeSelector`
4. Apply to all composition deployments

This ensures environment pods run on same node as user's WorkMachine for shared Nix store access.

### Status Updates
Tracks:
- `servicesCount`: Total services
- `runningCount`: Services with ready pods
- `services`: Individual service statuses
- `endpoints`: Service URLs
- `state`: deploying, deployed, failed

---

## ServiceIntercept Controller

**File**: `internal/controllers/serviceintercept/serviceintercept_controller.go`

### What It Does
Routes traffic from production services to workspace pods using SOCAT.

### Intercept Flow
1. Validate service and workspace exist
2. Get original service pods
3. Create SOCAT pod with:
   - Original service's labels (so service routes to it)
   - SOCAT command: `TCP-LISTEN:8080,fork TCP:workspace-headless:3000`
4. Delete original service pods
5. Traffic flows: Service → SOCAT Pod → Workspace

### Teardown Flow
1. Delete SOCAT pod
2. Deployment recreates original pods automatically
3. Traffic routes back to original pods

---

## Common Patterns

### Reconciliation Loop
```
1. Get resource
2. Handle deletion (if deletionTimestamp set)
3. Ensure finalizer exists
4. Reconcile desired state
5. Update status
```

### Finalizers
Used for cleanup before resource deletion:
- Add on resource creation
- Run cleanup logic when `deletionTimestamp` is set
- Remove finalizer to allow deletion

### Owner References
Set on dependent resources for automatic cleanup:
- When parent deleted, Kubernetes deletes children
- Example: Workspace pod has workspace as owner

### Status Updates
Always separate from spec updates:
- Use `r.Status().Update()` for status changes
- Preserves optimistic concurrency control

---

## Controller Responsibilities

### Controllers Handle
- Resource lifecycle (create/delete pods)
- State reconciliation
- Status updates
- Finalizers and cleanup

### Controllers Don't Handle
- HTTP requests (handlers do this)
- Field validation (webhooks do this)
- Authorization (handlers do this)
- Default values (webhooks do this)

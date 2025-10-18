# Handlers

Handlers manage HTTP API endpoints, handling business logic, authorization, and multi-resource operations.

## Responsibilities

### Handlers Handle
- Application business logic (multi-resource constraints)
- Authorization and authentication (role-based access)
- Coordinating operations across multiple resources
- User-facing API contracts

### Handlers Don't Handle
- Field format validation (webhooks do this)
- Default value injection (webhooks do this)
- Resource lifecycle management (controllers do this)

---

## User Handlers

**File**: `internal/handlers/user_handlers.go`

### Authorization Model
- **SuperAdmin**: Full access to all operations
- **Admin**: Can manage regular users only
- **User**: No administrative operations

### Key Endpoints

**`POST /api/v1/users`** - CreateUser
- Authorization: SuperAdmin can create any role; Admin can only create "user" role
- **Business Logic**: Auto-creates WorkMachine via UserService if user has 'user' role
- Webhook validates email format, uniqueness, role enums

**`PUT /api/v1/users/:name`** - UpdateUser
- Authorization: Dual check (can modify user? can assign new roles?)
- **Business Logic**:
  - If 'user' role added: Auto-create WorkMachine
  - If 'user' role removed: Delete WorkMachine
- Preserves ResourceVersion for optimistic concurrency

**`DELETE /api/v1/users/:name`** - DeleteUser
- Authorization: Same as UpdateUser
- Controller handles cleanup of associated resources

**`POST /api/v1/users/:name/reset-password`**
- Validation: Minimum 8 characters
- Updates password securely via service

**`GET /api/v1/users`** - ListUsers
- Query params: `labelSelector`, `fieldSelector`, `limit`, `continue`
- Returns paginated results

---

## WorkMachine Handlers

**File**: `internal/handlers/workmachine_handlers.go`

### **CRITICAL BUSINESS RULE: ONE WORKMACHINE PER USER**

This constraint is enforced in handlers, NOT webhooks, because it's an application-level business rule.

### Key Endpoints

**`POST /api/v1/work-machines/my`** - CreateMyWorkMachine
- **Business Rule Enforcement**: Checks if user already has WorkMachine, returns 409 Conflict if exists
- Determines machine type:
  - Use specified type if provided
  - Otherwise use default machine type
  - Error if no default exists
- Creates with DesiredState "stopped"
- Webhook generates deterministic name from owner email

**`POST /api/v1/work-machines/my/start`** - StartMyWorkMachine
- Sets DesiredState to "starting"
- Controller creates/starts pod

**`POST /api/v1/work-machines/my/stop`** - StopMyWorkMachine
- Sets DesiredState to "stopping"
- Controller stops/deletes pod

**`DELETE /api/v1/work-machines/my`** - DeleteMyWorkMachine
- Webhook prevents deletion if machine is running

---

## MachineType Handlers

**File**: `internal/handlers/machinetype_handlers.go`

### Authorization
All create/update/delete operations require authentication. Admin check TODO.

### Key Endpoints

**`GET /api/v1/machine-types`** - ListMachineTypes
- Public endpoint
- Query params: `active=true`, `category=xxx`

**`POST /api/v1/machine-types`** - CreateMachineType (Admin)
- Webhook validates:
  - Name format (lowercase alphanumeric with hyphens)
  - DisplayName uniqueness
  - Only one default machine type
  - Valid CPU/memory quantities

**`DELETE /api/v1/machine-types/:name`** - DeleteMachineType (Admin)
- Webhook prevents deletion if WorkMachines use this type

**`PUT /api/v1/machine-types/:name/activate`** - ActivateMachineType (Admin)
- Makes machine type available for selection

**`PUT /api/v1/machine-types/:name/deactivate`** - DeactivateMachineType (Admin)
- Hides from selection (existing machines unaffected)

---

## Environment Handlers

**File**: `internal/handlers/environment_handlers.go`

### Key Endpoints

**`POST /api/v1/environments`** - CreateEnvironment
- Validates user exists by email from JWT
- Sets `CreatedBy` to authenticated user email
- Webhook handles namespace creation and labels

**`DELETE /api/v1/environments/:name`** - DeleteEnvironment
- **Business Rule**: Requires `force=true` query param for activated environments
- Controller cleans up namespace and all resources

**`POST /api/v1/environments/:name/activate`** - ActivateEnvironment
- Sets Activated to true
- Controller scales up all resources in namespace

**`POST /api/v1/environments/:name/deactivate`** - DeactivateEnvironment
- Sets Activated to false
- Controller scales down all deployments to 0 replicas

**`PATCH /api/v1/environments/:name`** - PatchEnvironment
- Partial updates: `activated`, `labels`, `annotations`

---

## Workspace Handlers

**File**: `internal/handlers/workspace_handlers.go`

### Key Endpoints

**`POST /api/v1/namespaces/:namespace/workspaces`** - CreateWorkspace
- Finds user's WorkMachine to determine target namespace
- Uses WorkMachine's namespace (or WorkMachine name if TargetNamespace not set)
- Sets owner to authenticated user
- Webhook sets default resources and image

**`POST /api/v1/namespaces/:namespace/workspaces/:name/suspend`** - SuspendWorkspace
- Sets state to "suspended"
- Controller stops pod (preserves data on PVC)

**`POST /api/v1/namespaces/:namespace/workspaces/:name/activate`** - ActivateWorkspace
- Sets state to "active"
- Controller starts pod and restores access

**`POST /api/v1/namespaces/:namespace/workspaces/:name/archive`** - ArchiveWorkspace
- Sets state to "archived"
- Controller permanently suspends workspace

**`GET /api/v1/namespaces/:namespace/workspaces/:name/metrics`** - GetMetrics
- Queries Kubernetes Metrics API for pod metrics
- Aggregates CPU/memory usage across containers
- Returns current consumption and limits

**`GET /api/v1/nodes/:nodeName/metrics`** - GetNodeMetrics
- Returns node-level metrics from Kubernetes Metrics API

---

## Composition Handlers

**File**: `internal/handlers/composition_handlers.go`

### Key Endpoints

**`POST /api/v1/namespaces/:namespace/compositions`** - CreateComposition
- Webhook validates:
  - DisplayName required (max 100 chars)
  - ComposeContent is valid YAML
  - ComposeFormat is valid version
  - EnvFrom references exist

**`PUT /api/v1/namespaces/:namespace/compositions/:name`** - UpdateComposition
- Updates composition spec
- Controller triggers redeployment if autoDeploy is true

**`GET /api/v1/namespaces/:namespace/compositions/:name/status`** - GetCompositionStatus
- Returns deployment status:
  - State, message
  - Services count, running count
  - Service statuses and endpoints
  - Last deployed time

---

## ServiceIntercept Handlers

**File**: `internal/handlers/serviceintercept_handlers.go`

### Key Endpoints

**`POST /api/v1/namespaces/:namespace/service-intercepts`** - CreateServiceIntercept
- Controller creates SOCAT pod with original service labels
- Controller deletes original service pods
- Routes traffic through SOCAT to workspace

**`DELETE /api/v1/namespaces/:namespace/service-intercepts/:name`** - DeleteServiceIntercept
- Controller removes SOCAT pod
- Controller restores original service pods
- Returns traffic to normal routing

**`POST /api/v1/namespaces/:namespace/service-intercepts/:name/deactivate`**
- Deletes the intercept (intercepts have no inactive state)

---

## Environment Config Handlers

**File**: `internal/handlers/environment_config_handlers.go`

### Legacy Config/Secret Management

**`PUT /api/v1/environments/:name/config`** - SetConfig
- Creates or updates `env-config` ConfigMap in environment namespace
- Stores key-value pairs as environment variables

**`PUT /api/v1/environments/:name/secret`** - SetSecret
- Creates or updates `env-secret` Secret in environment namespace

**`GET /api/v1/environments/:name/secret`** - GetSecret
- **Security**: Returns only keys, NOT values (prevents accidental exposure)

### File Management

**`PUT /api/v1/environments/:name/files/:filename`** - SetFile
- **Security**: Validates filename (rejects path traversal with ".." or "/")
- Creates ConfigMap named `env-file-{filename}`

**`GET /api/v1/environments/:name/files`** - ListFiles
- Lists ConfigMaps with `file-type: environment-file` label

### Unified EnvVar Management (New API)

**`GET /api/v1/environments/:name/envvars`** - GetEnvVars
- Merges `env-config` ConfigMap and `env-secret` Secret
- **Security**: Secret values returned as empty string

**`POST /api/v1/environments/:name/envvars`** - CreateEnvVar
- **Business Rule**: Checks key uniqueness across configs and secrets
- Returns 409 Conflict if key exists

**`PUT /api/v1/environments/:name/envvars`** - SetEnvVar
- Updates existing or creates new envvar

**`DELETE /api/v1/environments/:name/envvars/:key`** - DeleteEnvVar
- **Smart Deletion**: Searches both ConfigMap and Secret
- Deletes entire ConfigMap/Secret if becomes empty

---

## Common Patterns

### Authorization
- JWT-based authentication with role checking
- Context extraction: `middleware.GetUserFromContext(c)`
- Role hierarchy: SuperAdmin > Admin > User

### Webhook Reliance
- Handlers defer field validation to webhooks
- Focus on business logic and authorization
- Extract user-friendly errors from webhook responses

### Error Handling
- Specific messages for webhook validation failures
- Proper HTTP status codes (400, 401, 403, 404, 409, 500)

### Resource Management
- Get-Update pattern preserves ResourceVersion
- Optimistic concurrency control
- Handle not-found errors appropriately

### Business Logic Placement
- Multi-resource constraints in handlers
- Cross-resource validation
- Authorization before operations

### Security
- Sensitive data (secrets) never fully exposed
- Path traversal prevention in file operations
- Email-based authentication and user lookup

---

## Handler vs Webhook vs Controller

### Handlers
- Application business logic
- Authorization and authentication
- Multi-resource operations
- Business rules (e.g., "one user one workmachine")

### Webhooks
- Field-level validation
- Default value injection
- Format validation (email, name patterns)
- Single-resource constraints

### Controllers
- Resource lifecycle management
- State reconciliation
- Pod creation/deletion
- Status updates
- Finalizers and cleanup

---

## Separation Examples

### WorkMachine Creation

**Handler**: Checks "one workmachine per user" rule, determines default machine type, sets DesiredState

**Webhook**: Generates name from owner, validates machine type exists, adds labels

**Controller**: Creates/deletes pod based on DesiredState, updates status

### Environment Deletion

**Handler**: Requires force flag for activated environments, deletes resource

**Webhook**: None

**Controller**: Deletes namespace, cleans up resources, handles finalizers

### User Creation

**Handler**: Checks role permissions, auto-creates WorkMachine if user has 'user' role

**Webhook**: Validates email format/uniqueness, validates roles, sets defaults

**Controller**: Updates user status, manages password hashing, handles finalizers

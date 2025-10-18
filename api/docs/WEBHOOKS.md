# Webhooks

Webhooks validate and mutate resources before they're persisted. They run on CREATE, UPDATE, and DELETE operations.

## Webhook Flow

```
kubectl apply → Mutation Webhook → Validation Webhook → Persist to etcd → Controller Reconciles
```

---

## User Webhook

**File**: `internal/webhooks/user_webhook.go`

### Validation
- Email format and uniqueness
- Required fields (email, username)
- Valid role enums

### Mutation
- Sets default status if not provided
- Adds user labels

---

## Environment Webhook

**File**: `internal/webhooks/environment_webhook.go`

### Validation
- DisplayName length (max 100 chars)
- Namespace naming conventions
- Resource quota formats (CPU, memory)
- **Namespace existence** (CREATE only):
  - Rejects if namespace already exists
  - Checks if managed by another environment

### Mutation
**TargetNamespace generation** (if not provided):
- Pattern: `env-{environment-name}`
- Example: Environment `production` → `targetNamespace: env-production`

**Note**: Webhook only sets the field value; controller creates the actual namespace resource.

---

## MachineType Webhook

**File**: `internal/webhooks/machinetype_webhook_gin.go`

### Validation
#### CREATE/UPDATE
- Name format: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
- DisplayName required
- Category required (general, compute-optimized, memory-optimized, gpu, development)
- Valid CPU/memory/GPU quantities
- No duplicate display names among active types
- Only one default machine type allowed

#### DELETE
- Prevents deletion if WorkMachines are using this type

### Mutation
- Sets default category: `general`
- Sets default priority: `100`
- Adds labels:
  - `kloudlite.io/machine-type-category`
  - `kloudlite.io/machine-type-active` (true/false)

---

## WorkMachine Webhook

**File**: `internal/webhooks/workmachine_webhook.go`

### Validation
- OwnedBy required
- Owner exists (email or username lookup)
- Machine type exists and is active
- **DELETE**: Prevents deleting running machines

**Note**: "One workmachine per user" rule is in handlers, not here.

### Mutation
**Name generation** (if not provided):
- Sanitizes owner: `john@example.com` → `wm-john-at-example-com`
- Replaces `@` with `-at-`, `.` with `-`

**Labels**:
- `kloudlite.io/owned-by`: User ID
- `kloudlite.io/owner-email`: Base64 encoded email
- `kloudlite.io/machine-type`: Machine type name

**NodeSelector** (if not provided):
- Automatically sets: `kloudlite.io/workmachine: <workmachine-name>`
- Ensures all user resources run on same node for shared Nix store access
- Propagated to: host-manager, workspaces, environment deployments

---

## Workspace Webhook

**File**: `internal/webhooks/workspace_webhook.go`

### Validation
- DisplayName and description length
- Valid resource quantities (requests/limits)
- Image configuration valid
- Environment references exist

### Mutation
- Sets default resource requests/limits
- Sets default image if not provided
- Adds workspace labels

---

## Composition Webhook

**File**: `internal/webhooks/composition_webhook.go`

### Validation
- DisplayName required (max 100 chars)
- Description optional (max 500 chars)
- ComposeContent required and valid YAML
- ComposeFormat valid (v2, v3, v3.1-v3.9)
- EnvFrom type: ConfigMap or Secret
- EnvFrom resources exist (warns if missing)
- ResourceOverrides: replicas 0-10

### Mutation
- Sets default ComposeFormat: `v3.8`

---

## ServiceIntercept Webhook

**File**: `internal/webhooks/serviceintercept_webhook.go`

### Validation
- Service exists
- Workspace exists
- Port mappings valid

### Mutation
- Adds intercept labels
- Sets default port mappings

---

## Webhook Responsibilities

### Webhooks Handle
- Field validation
- Format checking (regex, email)
- Resource existence checks
- Default value injection
- Label generation

### Webhooks Don't Handle
- Application business rules (handlers do this)
- Authorization (handlers do this)
- Multi-resource logic (handlers do this)
- Pod creation (controllers do this)

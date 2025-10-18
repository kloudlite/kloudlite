# Kloudlite API Documentation

Documentation for the Kloudlite API architecture and components.

## Documentation Files

### [CONTROLLERS.md](./CONTROLLERS.md)
Kubernetes controllers and reconciliation loops:
- User, Environment, WorkMachine, Workspace, Composition, ServiceIntercept controllers
- Reconciliation patterns, finalizers, status updates
- Resource lifecycle management

### [WEBHOOKS.md](./WEBHOOKS.md)
Kubernetes admission webhooks:
- Validation rules and mutation patterns
- Default value injection and label generation
- Field-level validation

### [HANDLERS.md](./HANDLERS.md)
HTTP API handlers:
- Authorization and business logic
- Multi-resource operations
- User-facing API endpoints

---

## Architecture Flow

```
HTTP Request
    ↓
HANDLERS (business logic, authorization)
    ↓
WEBHOOKS (validation, defaults)
    ↓
KUBERNETES API (persist resource)
    ↓
CONTROLLERS (reconcile state, manage pods)
```

---

## Separation of Concerns

### Handlers
- Application business logic
- Authorization (role-based)
- Multi-resource operations
- Business rules (e.g., "one user one workmachine")

### Webhooks
- Field-level validation
- Default value injection
- Format validation (regex, email)
- Label/name generation

### Controllers
- Resource lifecycle management
- State reconciliation
- Pod creation/deletion
- Status updates

---

## Key Patterns

### Business Rules in Handlers
Application-level constraints enforced in handlers, not webhooks.

**Example**: "One user one workmachine" rule in `workmachine_handlers.go:92-98`

### Validation in Webhooks
Field-level validation for consistency across all entry points (API, kubectl).

**Example**: Email validation, machine type validation in webhooks

### Lifecycle in Controllers
Resource lifecycle managed by controllers watching for changes.

**Example**: WorkMachine controller creates/deletes pods based on `DesiredState`

---

## Examples

### WorkMachine Creation

**Handler**: Checks "one per user" rule, determines default machine type, creates resource

**Webhook**: Generates name from owner, validates machine type exists, adds labels

**Controller**: Creates/deletes pod based on DesiredState, updates status

### Environment Deletion

**Handler**: Requires `force=true` for activated environments, deletes resource

**Webhook**: None

**Controller**: Deletes namespace, cleans up resources, handles finalizers

### User Creation

**Handler**: Checks role permissions, auto-creates WorkMachine if user has 'user' role

**Webhook**: Validates email/uniqueness, validates roles, sets defaults

**Controller**: Updates status, manages password hashing, handles finalizers

---

## Development Guidelines

### Adding a New Feature

1. Define CRD (Custom Resource Definition)
2. Implement webhook (validation/mutation)
3. Implement handler (API endpoint + business logic)
4. Implement controller (reconciliation logic)
5. Add tests

### Testing

```bash
# Run tests
go test ./internal/handlers/... -v
go test ./internal/webhooks/... -v
go test ./internal/controllers/... -v
```

### Debugging

```bash
# View API logs
kubectl logs -n default -l app=kloudlite-api

# View webhook errors
kubectl get events --field-selector reason=FailedCreate

# Force reconciliation
kubectl annotate resource/name force-reconcile=true --overwrite
```

---

## Resources

- **Kubernetes Documentation**: https://kubernetes.io/docs/
- **Controller Runtime**: https://github.com/kubernetes-sigs/controller-runtime
- **Admission Webhooks**: https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/

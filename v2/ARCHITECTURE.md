# Kloudlite v2 Architecture Documentation

## Table of Contents
1. [Project Overview](#project-overview)
2. [Project Structure](#project-structure)
3. [Architecture Patterns](#architecture-patterns)
4. [Backend Architecture](#backend-architecture)
5. [Frontend Architecture](#frontend-architecture)
6. [Custom Resource Definitions (CRDs)](#custom-resource-definitions-crds)
7. [API Design](#api-design)
8. [Security Architecture](#security-architecture)
9. [Development Workflow](#development-workflow)
10. [Deployment Architecture](#deployment-architecture)

## Project Overview

Kloudlite v2 is a Kubernetes-native platform that provides environment management, user management, and infrastructure orchestration capabilities. The system follows cloud-native principles with a strong emphasis on Kubernetes CRDs, clean architecture, and modern web development patterns.

### Key Components
- **Backend API Server**: Go-based REST API with Kubernetes integration
- **Frontend Web App**: Next.js 14 with TypeScript and Server Components
- **Kubernetes CRDs**: Custom resources for domain objects
- **Admission Webhooks**: Validation and mutation logic for CRDs
- **K3s Cluster**: Lightweight Kubernetes for development

## Project Structure

```
v2/
├── api/                        # Backend API Server
│   ├── cmd/
│   │   └── server/            # Main server entry point
│   │       └── main.go
│   ├── internal/
│   │   ├── server/            # HTTP server setup
│   │   │   └── server.go
│   │   ├── handlers/          # HTTP request handlers
│   │   │   ├── user_handlers.go
│   │   │   └── environment_handlers.go
│   │   ├── repository/        # Data access layer
│   │   │   ├── k8s_repository.go
│   │   │   ├── user_repository.go
│   │   │   └── environment_repository.go
│   │   ├── webhooks/          # Admission webhooks
│   │   │   ├── user_webhook.go
│   │   │   └── environment_webhook.go
│   │   ├── managers/          # Service managers
│   │   │   └── manager.go
│   │   └── middleware/        # HTTP middleware
│   │       └── auth.go
│   ├── pkg/
│   │   └── apis/             # CRD type definitions
│   │       ├── platform/v1alpha1/
│   │       │   ├── user_types.go
│   │       │   ├── groupversion_info.go
│   │       │   └── zz_generated.deepcopy.go
│   │       └── environments/v1/
│   │           ├── types.go
│   │           ├── register.go
│   │           └── zz_generated.deepcopy.go
│   ├── config/
│   │   └── crd/              # CRD YAML manifests
│   │       ├── bases/
│   │       └── rbac/
│   ├── kubeconfig/           # Generated kubeconfig files
│   └── Taskfile.yml         # Task automation
│
├── web/                      # Frontend Application
│   ├── src/
│   │   ├── app/             # Next.js App Router
│   │   │   ├── (dashboard)/
│   │   │   │   ├── layout.tsx
│   │   │   │   ├── users/
│   │   │   │   └── environments/
│   │   │   ├── actions/     # Server Actions
│   │   │   │   ├── user.actions.ts
│   │   │   │   └── environment.actions.ts
│   │   │   └── api/         # API routes (if needed)
│   │   ├── components/
│   │   │   ├── ui/          # Shadcn UI components
│   │   │   ├── dialogs/     # Dialog components
│   │   │   │   ├── create-environment.tsx
│   │   │   │   └── delete-environment-confirm.tsx
│   │   │   └── lists/       # List components
│   │   │       └── environment-list.tsx
│   │   ├── lib/
│   │   │   ├── api/         # API client
│   │   │   │   └── client.ts
│   │   │   └── utils/       # Utility functions
│   │   ├── types/           # TypeScript types
│   │   │   ├── user.ts
│   │   │   └── environment.ts
│   │   └── services/        # Service layer
│   │       ├── user-service.ts
│   │       └── environment-service.ts
│   ├── public/
│   └── package.json
│
├── kubeconfig/              # Shared kubeconfig directory
├── scripts/                 # Automation scripts
└── docker-compose.yml      # Development infrastructure
```

## Architecture Patterns

### 1. Clean Architecture (Backend)
The backend follows Clean Architecture principles with clear separation of concerns:

```
┌─────────────────────────────────────────────────┐
│                   HTTP Layer                    │
│              (handlers, middleware)             │
├─────────────────────────────────────────────────┤
│                 Service Layer                   │
│              (managers, business logic)         │
├─────────────────────────────────────────────────┤
│               Repository Layer                  │
│            (data access abstraction)            │
├─────────────────────────────────────────────────┤
│                Infrastructure                   │
│          (Kubernetes API, Database)            │
└─────────────────────────────────────────────────┘
```

### 2. Repository Pattern
Generic repository interface for all CRDs:

```go
type Repository[T runtime.Object, TList runtime.Object] interface {
    Create(ctx context.Context, obj T) error
    Get(ctx context.Context, name string) (T, error)
    Update(ctx context.Context, obj T) error
    Delete(ctx context.Context, name string) error
    List(ctx context.Context) (TList, error)
}
```

### 3. Service Manager Pattern
Centralized service management with dependency injection:

```go
type Manager struct {
    K8sClient            client.Client
    UserRepository       repository.UserRepository
    EnvironmentRepository repository.EnvironmentRepository
    UserWebhook         *webhooks.UserWebhook
    EnvironmentWebhook  *webhooks.EnvironmentWebhook
}
```

## Backend Architecture

### Server Setup
- **Framework**: Gin web framework for HTTP routing
- **Port**: 8080 (configurable via PORT env var)
- **Kubernetes Client**: controller-runtime client
- **Middleware**: Authentication, CORS, logging

### Request Flow
1. HTTP request → Gin router
2. Authentication middleware validates user
3. Handler processes request with business logic
4. Repository interacts with Kubernetes API
5. Response returned to client

### Webhook Architecture
Admission webhooks provide validation and mutation for CRDs:

```go
// Validation webhook
func (w *EnvironmentWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) error {
    env := obj.(*environmentsv1.Environment)
    // Validation logic
    return nil
}

// Mutation webhook
func (w *EnvironmentWebhook) Default(ctx context.Context, obj runtime.Object) error {
    env := obj.(*environmentsv1.Environment)
    // Mutation logic (e.g., adding labels)
    return nil
}
```

### Error Handling
- Structured error responses with status codes
- Validation errors return 400 Bad Request
- Not found errors return 404
- Server errors return 500 with sanitized messages

## Frontend Architecture

### Next.js 14 App Router
- **Server Components**: Default for all pages
- **Client Components**: Used only when necessary (forms, interactions)
- **Server Actions**: All API calls are server-side
- **TypeScript**: Strict type safety throughout

### Server Actions Pattern
All API interactions use server actions to maintain security:

```typescript
// app/actions/environment.actions.ts
'use server'

export async function createEnvironment(data: EnvironmentCreateRequest, user: string) {
    const result = await environmentService.createEnvironment(data, user)
    revalidatePath('/environments')
    return { success: true, data: result }
}
```

### Component Architecture
```
┌─────────────────────────────────────────────────┐
│              Server Components                  │
│          (Pages, Layouts, Data Fetching)       │
├─────────────────────────────────────────────────┤
│              Client Components                  │
│         (Forms, Dialogs, Interactions)         │
├─────────────────────────────────────────────────┤
│               Server Actions                    │
│            (API calls, mutations)              │
├─────────────────────────────────────────────────┤
│                 Services                        │
│          (Business logic, API client)          │
└─────────────────────────────────────────────────┘
```

### State Management
- Server state via React Server Components
- Client state via React hooks (useState, useReducer)
- Form state via controlled components
- Cache invalidation via revalidatePath

## Custom Resource Definitions (CRDs)

### Design Principles
1. **Namespace Scoping**: Resources can be cluster or namespace scoped
2. **Status Subresource**: Separate status updates from spec changes
3. **Validation**: OpenAPI schema validation in CRD
4. **Webhooks**: Additional validation and mutation logic
5. **Labels/Annotations**: Consistent metadata patterns

### User CRD
```yaml
apiVersion: platform.kloudlite.io/v1alpha1
kind: User
metadata:
  name: user-uuid
  labels:
    kloudlite.io/username: johndoe
    kloudlite.io/email: john@example.com
spec:
  username: johndoe
  email: john@example.com
  fullName: John Doe
  roles: [admin]
status:
  active: true
  lastLogin: "2024-09-28T10:00:00Z"
```

### Environment CRD
```yaml
apiVersion: environments.kloudlite.io/v1
kind: Environment
metadata:
  name: dev-environment
  labels:
    kloudlite.io/owned-by: user-uuid
    kloudlite.io/owner-email: am9obkBleGFtcGxlLmNvbQ==  # base64 encoded
spec:
  targetNamespace: env-dev-environment
  createdBy: johndoe
  activated: true
  resourceQuotas:
    limits.cpu: "4"
    limits.memory: "8Gi"
status:
  state: active
  lastActivatedTime: "2024-09-28T10:00:00Z"
```

### CRD Patterns

#### Ownership Pattern
All user-created resources include ownership metadata:
```go
labels["kloudlite.io/owned-by"] = userID
labels["kloudlite.io/owner-email"] = base64(email)
```

#### Activation Pattern
Resources can be activated/deactivated:
```go
type EnvironmentSpec struct {
    Activated bool `json:"activated"`
}
```

#### Validation Pattern
Webhooks validate business rules:
- User exists and is valid
- Names follow Kubernetes conventions
- Resource quotas are within limits

## API Design

### RESTful Endpoints
```
GET    /api/v1/users           # List users
POST   /api/v1/users           # Create user
GET    /api/v1/users/:id       # Get user
PUT    /api/v1/users/:id       # Update user
DELETE /api/v1/users/:id       # Delete user

GET    /api/v1/environments    # List environments
POST   /api/v1/environments    # Create environment
GET    /api/v1/environments/:name  # Get environment
PUT    /api/v1/environments/:name  # Update environment
DELETE /api/v1/environments/:name  # Delete environment
POST   /api/v1/environments/:name/activate    # Activate
POST   /api/v1/environments/:name/deactivate  # Deactivate
```

### Request/Response Format
```typescript
// Request
interface EnvironmentCreateRequest {
  name: string
  spec: {
    targetNamespace: string
    activated: boolean
    resourceQuotas?: ResourceQuotas
  }
}

// Response
interface ApiResponse<T> {
  success: boolean
  data?: T
  error?: string
}
```

### Authentication
- Header: `X-User-Name` or `X-User-Email`
- Future: JWT tokens with proper auth service

## Security Architecture

### Authentication & Authorization
1. **User Identification**: Currently via headers, planned JWT
2. **RBAC**: Kubernetes RBAC for service accounts
3. **Ownership Validation**: Resources tied to valid users
4. **Webhook Validation**: Business rule enforcement

### Data Security
1. **Sensitive Data**: Stored as base64 in labels (emails)
2. **Secrets Management**: Kubernetes secrets for sensitive config
3. **Input Validation**: At API and CRD levels
4. **SQL Injection**: N/A (Kubernetes API)
5. **XSS Prevention**: React auto-escaping, server-side rendering

### Network Security
1. **CORS**: Configured for frontend origin
2. **HTTPS**: Required in production
3. **Network Policies**: Kubernetes network policies
4. **Service Mesh**: Optional Istio/Linkerd integration

## Development Workflow

### Local Development Setup
```bash
# 1. Start infrastructure
docker-compose up -d

# 2. Apply CRDs
cd v2/api
task apply-crds

# 3. Run backend
task run

# 4. Run frontend
cd ../web
pnpm dev
```

### Testing Strategy
1. **Unit Tests**: Repository and handler logic
2. **Integration Tests**: API endpoints with test K8s
3. **E2E Tests**: Frontend workflows
4. **Webhook Tests**: Validation and mutation logic

### Code Generation
```bash
# Generate CRD manifests
controller-gen crd paths=./pkg/apis/... output:dir=./config/crd/bases

# Generate DeepCopy methods
controller-gen object paths=./pkg/apis/...
```

### Git Workflow
1. Feature branches from `development`
2. PR to `development` for review
3. Merge to `master` for release
4. Tag releases with semver

## Deployment Architecture

### Development Environment
```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Frontend   │────▶│   Backend    │────▶│     K3s      │
│   (local)    │     │   (local)    │     │   (docker)   │
└──────────────┘     └──────────────┘     └──────────────┘
```

### Production Architecture
```
┌──────────────┐
│  Ingress     │
└──────┬───────┘
        │
┌───────▼──────────────────────┐
│       Kubernetes Cluster      │
├───────────────┬───────────────┤
│   Frontend    │   Backend     │
│   (Next.js)   │   (Go API)    │
│   Deployment  │   Deployment  │
└───────────────┴───────────────┘
        │               │
        └───────┬───────┘
                │
        ┌───────▼───────┐
        │     CRDs      │
        │  (User, Env)  │
        └───────────────┘
```

### Kubernetes Resources
1. **Deployments**: Frontend and backend apps
2. **Services**: ClusterIP for internal, LoadBalancer for external
3. **ConfigMaps**: Application configuration
4. **Secrets**: Sensitive configuration
5. **RBAC**: ServiceAccounts with appropriate permissions
6. **CRDs**: Custom resources for domain objects
7. **Webhooks**: ValidatingWebhookConfiguration, MutatingWebhookConfiguration

### Scaling Considerations
1. **Horizontal Pod Autoscaling**: Based on CPU/memory
2. **Database**: Consider external database for state
3. **Caching**: Redis for session/cache data
4. **Message Queue**: NATS/RabbitMQ for async operations
5. **Multi-tenancy**: Namespace isolation per tenant

## Future Enhancements

### Planned Features
1. **WorkMachine CRD**: User-specific compute resources
2. **MachineType CRD**: Predefined machine configurations
3. **JWT Authentication**: Proper auth service with tokens
4. **Audit Logging**: Track all resource changes
5. **Metrics & Monitoring**: Prometheus integration
6. **Backup & Restore**: Velero integration
7. **CI/CD Integration**: GitOps workflows
8. **Multi-cluster Support**: Fleet management

### Technical Debt
1. Replace header-based auth with JWT
2. Add comprehensive error handling
3. Implement rate limiting
4. Add request/response logging
5. Improve webhook error messages
6. Add health check endpoints
7. Implement graceful shutdown

## Conclusion

Kloudlite v2 represents a modern, Kubernetes-native platform built with clean architecture principles. The separation of concerns, use of CRDs, and modern web development patterns provide a solid foundation for future growth and scalability.

The architecture prioritizes:
- **Maintainability**: Clear structure and patterns
- **Scalability**: Kubernetes-native design
- **Security**: Multiple layers of validation
- **Developer Experience**: Modern tooling and workflows
- **Extensibility**: Easy to add new CRDs and features

This document serves as the definitive guide for understanding and extending the Kloudlite v2 platform.
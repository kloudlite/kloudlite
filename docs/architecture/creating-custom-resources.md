# Creating Custom Resources in Kloudlite

This guide provides a complete checklist for implementing new Kubernetes Custom Resources (CRDs) in the Kloudlite platform.

## Overview

Custom Resources extend Kubernetes with new resource types. In Kloudlite, we use CRDs for domain objects like MachineType, WorkMachine, Environment, and User.

## Architecture Flow

```
Frontend (React/Next.js)
    ↓ Server Actions
    ↓ Service Layer
    ↓ API Client
    ↓ HTTP Request
Backend API (Gin)
    ↓ Handlers (Auth only)
    ↓ Webhooks (Validation)
    ↓ Repository
    ↓ K8s Client
Kubernetes API
    ↓ CRD Storage
```

## Step-by-Step Implementation Checklist

### 1. Define the CRD Types (`/api/pkg/apis/{group}/v1/`)

#### a. Create Type Definitions
```go
// File: /api/pkg/apis/machines/v1/machinetype_types.go

package v1

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced  // For cluster-scoped resources
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster  // Or scope=Namespaced
// +kubebuilder:printcolumn:name="Name",type=string,JSONPath=`.metadata.name`
type MachineType struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   MachineTypeSpec   `json:"spec,omitempty"`
    Status MachineTypeStatus `json:"status,omitempty"`
}

type MachineTypeSpec struct {
    // +kubebuilder:validation:Required
    DisplayName string `json:"displayName"`

    // +kubebuilder:validation:Enum=general;compute-optimized;memory-optimized;gpu
    // +kubebuilder:default=general
    Category string `json:"category"`

    // +kubebuilder:validation:Required
    Resources MachineResources `json:"resources"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type MachineTypeList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items []MachineType `json:"items"`
}
```

**Checklist:**
- [ ] Add kubebuilder markers for code generation
- [ ] Define Spec with validation markers
- [ ] Define Status for observed state
- [ ] Create List type
- [ ] Add print columns for kubectl output
- [ ] Specify scope (Cluster or Namespaced)

### 2. Generate Code

```bash
# Run code generation
cd /api
go generate ./...

# Or manually:
controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./pkg/apis/..."
```

**Generated files:**
- `zz_generated.deepcopy.go` - DeepCopy methods
- CRD YAML manifests (if using controller-gen with crd output)

### 3. Create Repository (`/api/internal/repository/`)

```go
// File: /api/internal/repository/machinetype_repository.go

package repository

import (
    machinesv1 "github.com/kloudlite/kloudlite/v2/api/pkg/apis/machines/v1"
)

type MachineTypeRepository struct {
    *K8sRepository[*machinesv1.MachineType, *machinesv1.MachineTypeList]
}

func NewMachineTypeRepository(k8sClient client.Client) *MachineTypeRepository {
    return &MachineTypeRepository{
        K8sRepository: NewK8sRepository[*machinesv1.MachineType, *machinesv1.MachineTypeList](
            k8sClient,
            func() *machinesv1.MachineType { return &machinesv1.MachineType{} },
            func() *machinesv1.MachineTypeList { return &machinesv1.MachineTypeList{} },
        ),
    }
}

// Add custom methods if needed
func (r *MachineTypeRepository) ListActive(ctx context.Context) (*machinesv1.MachineTypeList, error) {
    list := &machinesv1.MachineTypeList{}
    err := r.k8sClient.List(ctx, list, client.MatchingLabels{
        "kloudlite.io/machine-type-active": "true",
    })
    return list, err
}
```

**Checklist:**
- [ ] Extend generic K8sRepository
- [ ] Implement constructor with proper types
- [ ] Add domain-specific methods
- [ ] Handle cluster-scoped vs namespaced resources

### 4. Create Admission Webhooks (`/api/internal/webhooks/`)

```go
// File: /api/internal/webhooks/machinetype_webhook.go

package webhooks

type MachineTypeWebhook struct {
    k8sClient client.Client
}

// Default - Set defaults and labels
func (w *MachineTypeWebhook) Default(ctx context.Context, obj runtime.Object) error {
    machineType := obj.(*machinesv1.MachineType)

    // Set defaults
    if machineType.Spec.Category == "" {
        machineType.Spec.Category = "general"
    }

    // Set labels
    if machineType.Labels == nil {
        machineType.Labels = make(map[string]string)
    }
    machineType.Labels["kloudlite.io/category"] = machineType.Spec.Category

    return nil
}

// ValidateCreate - ALL business validation here
func (w *MachineTypeWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) error {
    machineType := obj.(*machinesv1.MachineType)

    // Validate required fields
    if machineType.Spec.DisplayName == "" {
        return fmt.Errorf("displayName is required")
    }

    // Validate format
    if !isValidName(machineType.Name) {
        return fmt.Errorf("invalid name format")
    }

    // Validate business rules
    if err := w.checkDuplicates(ctx, machineType); err != nil {
        return err
    }

    return nil
}

// ValidateUpdate - Validate updates
func (w *MachineTypeWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) error {
    // Similar validation logic
    return nil
}

// ValidateDelete - Validate deletion
func (w *MachineTypeWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) error {
    // Check if resource is in use
    return nil
}
```

**Checklist:**
- [ ] Implement Default() for mutations
- [ ] Implement ValidateCreate() with ALL validations
- [ ] Implement ValidateUpdate()
- [ ] Implement ValidateDelete()
- [ ] NO business logic in handlers, only in webhooks

### 5. Create HTTP Handlers (`/api/internal/handlers/`)

```go
// File: /api/internal/handlers/machinetype_handlers.go

package handlers

type MachineTypeHandlers struct {
    manager *managers.Manager
}

func (h *MachineTypeHandlers) CreateMachineType(c *gin.Context) {
    ctx := c.Request.Context()

    // ONLY Authentication check
    userName := c.GetHeader("X-User-Email")
    if userName == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    // Parse request
    var req MachineTypeCreateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Build resource
    machineType := &machinesv1.MachineType{
        ObjectMeta: metav1.ObjectMeta{Name: req.Name},
        Spec: req.Spec,
    }

    // Apply defaults via webhook
    if err := h.manager.MachineTypeWebhook.Default(ctx, machineType); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Validate via webhook - ALL validation happens here
    if err := h.manager.MachineTypeWebhook.ValidateCreate(ctx, machineType); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Create resource
    if err := h.manager.MachineTypeRepository.Create(ctx, machineType); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, machineType)
}
```

**Checklist:**
- [ ] Authentication check ONLY (no other validation)
- [ ] Parse request body
- [ ] Call webhook.Default()
- [ ] Call webhook.ValidateCreate/Update/Delete()
- [ ] Call repository methods
- [ ] Return appropriate HTTP responses

### 6. Register Routes (`/api/internal/server/routes.go`)

```go
func setupRouter(cfg *config.Config, logger *zap.Logger, servicesManager *services.Manager) *gin.Engine {
    // Create manager with repositories and webhooks
    manager := &managers.Manager{
        K8sClient:             servicesManager.RepositoryManager.K8sClient,
        MachineTypeRepository: servicesManager.RepositoryManager.MachineTypes,
        MachineTypeWebhook:   webhooks.NewMachineTypeWebhook(servicesManager.RepositoryManager.K8sClient),
    }

    // Create handlers
    machineTypeHandlers := handlers.NewMachineTypeHandlers(manager)

    // Register routes
    v1 := router.Group("/api/v1")
    {
        machineTypes := v1.Group("/machine-types")
        {
            machineTypes.GET("", machineTypeHandlers.ListMachineTypes)
            machineTypes.POST("", machineTypeHandlers.CreateMachineType)
            machineTypes.GET("/:name", machineTypeHandlers.GetMachineType)
            machineTypes.PUT("/:name", machineTypeHandlers.UpdateMachineType)
            machineTypes.DELETE("/:name", machineTypeHandlers.DeleteMachineType)

            // Custom actions
            machineTypes.PUT("/:name/activate", machineTypeHandlers.ActivateMachineType)
            machineTypes.PUT("/:name/deactivate", machineTypeHandlers.DeactivateMachineType)
        }
    }
}
```

### 7. Frontend Types (`/web/src/types/`)

```typescript
// File: /web/src/types/machine.ts

export interface MachineType {
  metadata: {
    name: string
    creationTimestamp?: string
  }
  spec: MachineTypeSpec
  status?: MachineTypeStatus
}

export interface MachineTypeSpec {
  displayName: string
  category: 'general' | 'compute-optimized' | 'memory-optimized' | 'gpu'
  resources: {
    cpu: string
    memory: string
    storage: string
    gpu?: string
  }
  active?: boolean
}

export interface MachineTypeCreateRequest {
  name: string
  displayName: string
  category: 'general' | 'compute-optimized' | 'memory-optimized' | 'gpu'
  cpu: number
  memory: number
  storage: number
  gpu?: number
}
```

### 8. Frontend Service (`/web/src/lib/services/`)

```typescript
// File: /web/src/lib/services/machine-type.service.ts

export class MachineTypeService {
  private baseUrl = '/api/v1/machine-types'

  async listMachineTypes(user?: string): Promise<MachineTypeListResponse> {
    const headers: Record<string, string> = {}
    if (user) {
      headers['X-User-Email'] = user
    }
    return apiClient.get<MachineTypeListResponse>(this.baseUrl, { headers })
  }

  async createMachineType(data: MachineTypeCreateRequest, user?: string): Promise<MachineType> {
    const headers: Record<string, string> = {}
    if (user) {
      headers['X-User-Email'] = user
    }
    return apiClient.post<MachineType>(this.baseUrl, data, { headers })
  }
}

export const machineTypeService = new MachineTypeService()
```

### 9. Server Actions (`/web/src/app/actions/`)

```typescript
// File: /web/src/app/actions/machine-type.actions.ts

'use server'

import { auth } from '@/lib/auth'
import { machineTypeService } from '@/lib/services/machine-type.service'
import { revalidatePath } from 'next/cache'

export async function listMachineTypes() {
  try {
    // Get session for authentication
    const session = await auth()
    const userEmail = session?.user?.email

    const result = await machineTypeService.listMachineTypes(userEmail)
    return { success: true, data: result }
  } catch (error: any) {
    return {
      success: false,
      error: error.message || 'Failed to list machine types'
    }
  }
}

export async function createMachineType(data: MachineTypeCreateRequest) {
  try {
    const session = await auth()
    const userEmail = session?.user?.email

    const result = await machineTypeService.createMachineType(data, userEmail)
    revalidatePath('/admin/machine-configs')
    return { success: true, data: result }
  } catch (error: any) {
    return {
      success: false,
      error: error.message || 'Failed to create machine type'
    }
  }
}
```

### 10. React Components (`/web/src/components/`)

```typescript
// File: /web/src/components/machine-configs-list.tsx

'use client'

import { useState, useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { createMachineType, deleteMachineType } from '@/app/actions/machine-type.actions'

export function MachineConfigsList({ configs }: { configs: MachineConfig[] }) {
  const router = useRouter()
  const [isPending, startTransition] = useTransition()

  const handleCreate = async (data: FormData) => {
    const result = await createMachineType({
      name: data.get('name') as string,
      displayName: data.get('displayName') as string,
      category: data.get('category') as string,
      // ... other fields
    })

    if (result.success) {
      toast.success('Machine type created')
      startTransition(() => {
        router.refresh()
      })
    } else {
      toast.error(result.error)
    }
  }

  // Component JSX
}
```

### 11. Page Component (`/web/src/app/`)

```typescript
// File: /web/src/app/admin/machine-configs/page.tsx

import { listMachineTypes } from '@/app/actions/machine-type.actions'
import { MachineConfigsList } from '@/components/machine-configs-list'

export default async function MachineConfigsPage() {
  // Fetch data server-side
  const result = await listMachineTypes()

  const machineTypes = result.success && result.data
    ? result.data.items || []
    : []

  // Transform data if needed
  const transformedConfigs = transformMachineTypes(machineTypes)

  return (
    <main className="space-y-6">
      <h1 className="text-3xl font-light">Machine Configurations</h1>
      <MachineConfigsList configs={transformedConfigs} />
    </main>
  )
}
```

## Testing Checklist

### Backend Testing

```bash
# 1. Test CRD creation
kubectl apply -f config/crd/machinetype.yaml

# 2. Test CRUD operations via API
curl -X POST localhost:8080/api/v1/machine-types \
  -H "X-User-Email: admin@example.com" \
  -H "Content-Type: application/json" \
  -d '{"name":"test","spec":{...}}'

# 3. Verify webhook validation
# Try creating without required fields
curl -X POST localhost:8080/api/v1/machine-types \
  -H "X-User-Email: admin@example.com" \
  -H "Content-Type: application/json" \
  -d '{"name":"test","spec":{}}'
# Should return validation error

# 4. Check Kubernetes resources
kubectl get machinetypes
kubectl describe machinetype test
```

### Frontend Testing

```bash
# 1. Test page loads
npm run dev
# Navigate to /admin/machine-configs

# 2. Test CRUD operations
# - Create new machine type
# - Edit existing
# - Delete
# - Activate/Deactivate

# 3. Verify validation
# - Try submitting without required fields
# - Check category dropdown values
# - Verify error messages
```

## Common Patterns

### 1. Cluster-scoped vs Namespaced Resources

```go
// Cluster-scoped (like MachineType)
// +genclient:nonNamespaced
// +kubebuilder:resource:scope=Cluster

// Namespaced (like WorkMachine)
// +kubebuilder:resource:scope=Namespaced
```

### 2. Owner References

```go
// Set owner for garbage collection
machineType.SetOwnerReferences([]metav1.OwnerReference{{
    APIVersion: "v1",
    Kind:       "User",
    Name:       userName,
    UID:        userUID,
}})
```

### 3. Labels for Filtering

```go
// Set labels in webhook Default()
obj.Labels["kloudlite.io/active"] = "true"
obj.Labels["kloudlite.io/category"] = obj.Spec.Category

// Query by labels
client.List(ctx, list, client.MatchingLabels{
    "kloudlite.io/active": "true",
})
```

### 4. Status Updates

```go
// Update status separately
machineType.Status.InUseCount = count
machineType.Status.LastUpdated = &metav1.Time{Time: time.Now()}
err := r.k8sClient.Status().Update(ctx, machineType)
```

## Validation Strategy

### What Goes Where:

**Webhook Validators:**
- ALL business logic validation
- Required field checks
- Format validation
- Uniqueness checks
- Referential integrity
- Resource limits
- State transition rules

**HTTP Handlers:**
- Authentication only (X-User-Email header)
- Authorization checks (admin role, ownership)
- NO other validation

**Frontend:**
- Client-side validation for UX
- Same rules as webhook for consistency
- Immediate feedback to user

## Troubleshooting

### Common Issues:

1. **CRD not found**: Run code generation, apply CRD YAML
2. **Validation not working**: Check webhook is called in handler
3. **404 on API calls**: Verify routes are registered
4. **Authentication errors**: Ensure session includes user email
5. **Kubernetes permission denied**: Check RBAC rules

### Debug Commands:

```bash
# Check CRD is installed
kubectl get crd machinetypes.machines.kloudlite.io

# Watch API logs
kubectl logs -f deployment/kloudlite-api

# Test webhook directly
kubectl create -f test-resource.yaml --dry-run=server

# Check generated code
ls -la pkg/apis/machines/v1/zz_generated*
```

## Best Practices

1. **Always validate in webhooks** - Single source of truth
2. **Keep handlers thin** - Only HTTP concerns
3. **Use generic repository** - Avoid duplication
4. **Set meaningful defaults** - In webhook Default()
5. **Use labels for queries** - Efficient filtering
6. **Handle errors gracefully** - Return clear messages
7. **Test webhook validation** - Unit tests for business rules
8. **Document API changes** - Update OpenAPI/Proto specs
9. **Version your APIs** - Use v1, v2 for breaking changes
10. **Monitor resource usage** - Set resource limits

## Summary

Creating a new Custom Resource involves:
1. Define types with validation markers
2. Generate DeepCopy code
3. Create repository using generic base
4. Implement webhooks with ALL validation
5. Create thin HTTP handlers (auth only)
6. Register API routes
7. Define TypeScript types
8. Create service layer
9. Implement server actions
10. Build React components
11. Test everything

Remember: **Webhooks own validation, Handlers own authentication**
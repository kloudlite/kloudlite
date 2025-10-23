// Machine Type interfaces based on the backend API
export interface MachineType {
  metadata: {
    name: string
    creationTimestamp?: string
    resourceVersion?: string
    uid?: string
    labels?: Record<string, string>
  }
  spec: MachineTypeSpec
  status?: MachineTypeStatus
}

export interface MachineTypeSpec {
  // Resources (can be either direct fields or in a resources object)
  cpu?: number | string
  memory?: number | string // in GB
  gpu?: number | string

  // K8s-style resources object
  resources?: {
    cpu?: string
    memory?: string
    gpu?: string
  }

  // Categorization
  category?: 'general' | 'compute-optimized' | 'memory-optimized' | 'gpu' | 'development'

  // Status
  active?: boolean

  // Display
  description?: string
  displayName?: string

  // K8s scheduling
  priority?: number
  nodeSelector?: Record<string, string>
  tolerations?: Array<{
    key?: string
    operator?: string
    value?: string
    effect?: string
  }>
}

export interface MachineTypeStatus {
  phase?: string
  conditions?: Array<{
    type: string
    status: string
    lastTransitionTime?: string
    reason?: string
    message?: string
  }>
}

// Request/Response types
export interface MachineTypeCreateRequest {
  name: string
  displayName?: string
  description?: string
  cpu: number
  memory: number
  gpu?: number
  category: 'general' | 'compute-optimized' | 'memory-optimized' | 'gpu' | 'development'
  active?: boolean
}

export interface MachineTypeUpdateRequest {
  displayName?: string
  description?: string
  cpu?: number
  memory?: number
  gpu?: number
  category?: 'general' | 'compute-optimized' | 'memory-optimized' | 'gpu' | 'development'
  active?: boolean
}

export interface MachineTypeListResponse {
  items: MachineType[]
  count: number
}

export interface MachineTypeResponse {
  success: boolean
  data?: MachineType
  error?: string
}

export interface MachineTypeDeleteResponse {
  success: boolean
  message?: string
  error?: string
}

// Composition types matching the backend API structure

export interface CompositionMetadata {
  name: string
  namespace: string
  uid?: string
  resourceVersion?: string
  generation?: number
  creationTimestamp?: string
  deletionTimestamp?: string
  managedFields?: any[]
}

export interface EnvFromSource {
  type: 'ConfigMap' | 'Secret'
  name: string
  prefix?: string
}

export interface ServiceResourceOverride {
  cpu?: string
  memory?: string
  replicas?: number
}

export interface CompositionSpec {
  displayName: string
  description?: string
  composeContent: string
  composeFormat?: string
  envVars?: Record<string, string>
  envFrom?: EnvFromSource[]
  autoDeploy?: boolean
  resourceOverrides?: Record<string, ServiceResourceOverride>
}

export interface ServiceStatus {
  name: string
  state: 'pending' | 'starting' | 'running' | 'stopped' | 'failed'
  replicas?: number
  readyReplicas?: number
  image?: string
  ports?: number[]
  message?: string
}

export interface DeployedResources {
  deployments?: string[]
  services?: string[]
  configMaps?: string[]
  secrets?: string[]
  pvcs?: string[]
  networkPolicies?: string[]
}

export interface CompositionStatus {
  state?: 'pending' | 'deploying' | 'running' | 'degraded' | 'stopped' | 'failed' | 'deleting'
  message?: string
  servicesCount?: number
  runningCount?: number
  services?: ServiceStatus[]
  endpoints?: Record<string, string>
  lastDeployedTime?: string
  conditions?: Array<{
    type: string
    status: string
    reason?: string
    message?: string
    lastTransitionTime?: string
  }>
  observedGeneration?: number
  deployedResources?: DeployedResources
}

export interface Composition {
  metadata: CompositionMetadata
  spec: CompositionSpec
  status?: CompositionStatus
}

// API Request types
export interface CompositionCreateRequest {
  name: string
  spec: CompositionSpec
}

export interface CompositionUpdateRequest {
  spec: CompositionSpec
}

// API Response types
export interface CompositionListResponse {
  compositions: Composition[]
  count: number
}

export interface CompositionResponse {
  message: string
  composition: Composition
}

export interface CompositionDeleteResponse {
  name: string
  namespace: string
  message: string
}

export interface CompositionStatusResponse {
  name: string
  namespace: string
  state?: string
  message?: string
  servicesCount?: number
  runningCount?: number
  services?: ServiceStatus[]
  endpoints?: Record<string, string>
  lastDeployedTime?: string
}

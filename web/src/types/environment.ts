// Environment types matching the backend API structure

export interface EnvironmentMetadata {
  name: string
  uid?: string
  resourceVersion?: string
  generation?: number
  creationTimestamp?: string
  deletionTimestamp?: string
  managedFields?: unknown[]
}

export interface ResourceQuotas {
  limitsCPU?: string
  limitsMemory?: string
  requestsCPU?: string
  requestsMemory?: string
  persistentVolumeClaims?: string
}

export interface NetworkPolicyPort {
  port: number
  protocol?: 'TCP' | 'UDP'
}

export interface LabelSelector {
  matchLabels?: Record<string, string>
}

export interface NetworkPolicyPeer {
  namespaceSelector?: LabelSelector
  podSelector?: LabelSelector
}

export interface IngressRule {
  from?: NetworkPolicyPeer[]
  ports?: NetworkPolicyPort[]
}

export interface NetworkPolicies {
  allowedNamespaces?: string[]
  ingressRules?: IngressRule[]
}

export interface EnvironmentSpec {
  targetNamespace?: string
  ownedBy: string
  activated: boolean
  labels?: Record<string, string>
  annotations?: Record<string, string>
  resourceQuotas?: ResourceQuotas
  networkPolicies?: NetworkPolicies
  cloneFrom?: string
}

export interface EnvironmentStatus {
  state?: 'active' | 'inactive' | 'activating' | 'deactivating' | 'deleting' | 'error'
  message?: string
  phase?: string
  namespaceCreated?: boolean
  resourcesApplied?: boolean
  lastActivatedTime?: string
  lastDeactivatedTime?: string
  resourceCount?: {
    deployments?: number
    services?: number
    configMaps?: number
    secrets?: number
  }
  conditions?: Array<{
    type: string
    status: string
    reason?: string
    message?: string
    lastTransitionTime?: string
  }>
}

export interface Environment {
  metadata: EnvironmentMetadata
  spec: EnvironmentSpec
  status?: EnvironmentStatus
}

// API Response types
export interface EnvironmentCreateRequest {
  name: string
  spec: EnvironmentSpec
}

export interface EnvironmentUpdateRequest {
  spec: EnvironmentSpec
}

export interface EnvironmentListResponse {
  environments: Environment[]
  count: number
}

export interface EnvironmentResponse {
  environment: Environment
  message: string
}

export interface EnvironmentDeleteResponse {
  name: string
  message: string
}

export interface EnvironmentStatusResponse {
  status: EnvironmentStatus
  message: string
}

// Config, Secret, and File management types
export interface ConfigData {
  [key: string]: string
}

export interface SetConfigRequest {
  data: ConfigData
}

export interface SetConfigResponse {
  message: string
  data: ConfigData
}

export interface GetConfigResponse {
  data: ConfigData
}

export interface DeleteConfigResponse {
  message: string
}

export interface SecretData {
  [key: string]: string
}

export interface SetSecretRequest {
  data: SecretData
}

export interface SetSecretResponse {
  message: string
  keys: string[]
}

export interface GetSecretResponse {
  keys: string[]
}

export interface DeleteSecretResponse {
  message: string
}

// EnvVars types (unified configs + secrets)
export interface EnvVar {
  key: string
  value: string // Empty for secrets (security)
  type: 'config' | 'secret'
}

export interface GetEnvVarsResponse {
  envVars: EnvVar[]
  count: number
}

export interface SetEnvVarRequest {
  key: string
  value: string
  type: 'config' | 'secret'
}

export interface SetEnvVarResponse {
  message: string
  envVar: EnvVar
}

export interface DeleteEnvVarResponse {
  message: string
}

export interface FileInfo {
  name: string
  configMapName: string
}

export interface SetFileRequest {
  content: string
}

export interface SetFileResponse {
  message: string
  filename: string
}

export interface GetFileResponse {
  filename: string
  content: string
}

export interface ListFilesResponse {
  files: FileInfo[]
  count: number
}

export interface DeleteFileResponse {
  message: string
}

// UI-specific types for compatibility with existing components
export interface EnvironmentUIModel {
  id: string
  name: string
  owner: string
  status: 'active' | 'inactive' | 'activating' | 'deactivating' | 'deleting' | 'error'
  created: string
  targetNamespace: string
  services: number
  configs: number
  secrets: number
  workspaces: string[]
  lastDeployed: string
}

// Converter functions
export function environmentToUIModel(env: Environment, owner?: string): EnvironmentUIModel {
  const createdDate = env.metadata.creationTimestamp
    ? new Date(env.metadata.creationTimestamp)
    : new Date()

  const now = new Date()
  const diffMs = now.getTime() - createdDate.getTime()
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))

  let createdText = 'Just now'
  if (diffDays > 30) {
    createdText = `${Math.floor(diffDays / 30)} month${Math.floor(diffDays / 30) > 1 ? 's' : ''} ago`
  } else if (diffDays > 7) {
    createdText = `${Math.floor(diffDays / 7)} week${Math.floor(diffDays / 7) > 1 ? 's' : ''} ago`
  } else if (diffDays > 0) {
    createdText = `${diffDays} day${diffDays > 1 ? 's' : ''} ago`
  }

  // Determine status: prioritize deletionTimestamp, then status.state, then spec.activated
  let status: 'active' | 'inactive' | 'activating' | 'deactivating' | 'deleting' | 'error'

  if (env.metadata.deletionTimestamp) {
    // If deletionTimestamp is set, the resource is being deleted
    status = 'deleting'
  } else if (env.status?.state) {
    // Use the controller-reported state if available
    status = env.status.state
  } else {
    // Fall back to spec.activated
    status = env.spec.activated ? 'active' : 'inactive'
  }

  return {
    id: env.metadata.name,
    name: env.metadata.name,
    owner: owner || env.spec.labels?.['kloudlite.io/owned-by'] || 'unknown',
    status,
    created: createdText,
    targetNamespace: env.spec.targetNamespace || '',
    services: env.status?.resourceCount?.services || 0,
    configs: env.status?.resourceCount?.configMaps || 0,
    secrets: env.status?.resourceCount?.secrets || 0,
    workspaces: [],
    lastDeployed: env.status?.lastActivatedTime
      ? new Date(env.status.lastActivatedTime).toLocaleString()
      : 'Never',
  }
}

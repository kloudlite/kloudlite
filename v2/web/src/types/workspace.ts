// Workspace interfaces based on the backend API
export interface Workspace {
  metadata: {
    name: string
    namespace: string
    creationTimestamp?: string
    resourceVersion?: string
    uid?: string
    labels?: Record<string, string>
    annotations?: Record<string, string>
  }
  spec: WorkspaceSpec
  status?: WorkspaceStatus
}

export interface WorkspaceSpec {
  displayName: string
  description?: string
  owner: string
  workMachineRef: ObjectReference
  environmentRef?: ObjectReference
  machineTypeRef?: ObjectReference
  resourceQuota?: ResourceQuota
  settings?: WorkspaceSettings
  status?: 'active' | 'suspended' | 'archived'
  tags?: string[]
}

export interface ObjectReference {
  name: string
  namespace: string
  kind?: string
  apiVersion?: string
}

export interface ResourceQuota {
  cpu?: string
  memory?: string
  storage?: string
  gpus?: number
}

export interface WorkspaceSettings {
  autoStop?: boolean
  idleTimeout?: number
  maxRuntime?: number
  startupScript?: string
  environmentVariables?: Record<string, string>
  vscodeExtensions?: string[]
  gitConfig?: {
    userName?: string
    userEmail?: string
    defaultBranch?: string
  }
  dotfilesRepo?: string
}

export interface WorkspaceStatus {
  phase?: string
  message?: string
  ready?: boolean
  conditions?: Array<{
    type: string
    status: string
    lastTransitionTime?: string
    reason?: string
    message?: string
  }>
  resources?: {
    cpu?: string
    memory?: string
    storage?: string
  }
  lastActivity?: string
}

// Request/Response types
export interface WorkspaceCreateRequest {
  name: string
  spec: WorkspaceSpec
}

export interface WorkspaceUpdateRequest {
  spec: WorkspaceSpec
}

export interface WorkspaceListResponse {
  items: Workspace[]
  metadata?: {
    continue?: string
    resourceVersion?: string
  }
}

export interface WorkspaceListParams {
  namespace?: string
  owner?: string
  workMachine?: string
  status?: 'active' | 'suspended' | 'archived'
  limit?: number
  continue?: string
}

export interface WorkspaceActionResponse {
  message: string
  error?: string
}
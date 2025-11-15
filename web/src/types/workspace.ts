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
  ownedBy: string
  workMachineRef?: ObjectReference
  workmachineName?: string
  environmentRef?: ObjectReference
  machineTypeRef?: ObjectReference
  folderName?: string
  packages?: PackageSpec[]
  resourceQuota?: ResourceQuota
  settings?: WorkspaceSettings
  status?: 'active' | 'suspended' | 'archived'
  tags?: string[]
  vscodeVersion?: string
  gitRepository?: GitRepository
  copyFrom?: string
}

export interface GitRepository {
  url: string
  branch?: string
}

export interface PackageSpec {
  name: string
  channel?: string
  nixpkgsCommit?: string
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

export interface InstalledPackage {
  name: string
  version?: string
  binPath?: string
  storePath?: string
  installedAt?: string
}

export interface ConnectedEnvironmentInfo {
  name: string
  targetNamespace: string
  availableServices?: string[]
}

export interface CloningStatus {
  phase?: string
  message?: string
  sourceWorkspaceName?: string
  sourceWorkmachineName?: string
  sourceFolderName?: string
  copyJobStatus?: {
    senderJobName?: string
    receiverJobName?: string
    senderPodIP?: string
    started?: boolean
    completed?: boolean
    failed?: boolean
    message?: string
  }
  startTime?: string
  completionTime?: string
  errorMessage?: string
}

export interface SourceCloningStatus {
  targetWorkspaceName?: string
  suspended?: boolean
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
  installedPackages?: InstalledPackage[]
  failedPackages?: string[]
  packageInstallationMessage?: string
  accessUrl?: string
  accessUrls?: Record<string, string> // Multiple access URLs for different services
  podName?: string
  podIP?: string
  nodeName?: string
  startTime?: string
  stopTime?: string
  totalRuntime?: number
  connectedEnvironment?: ConnectedEnvironmentInfo
  activeConnections?: number
  cloningStatus?: CloningStatus
  sourceCloningStatus?: SourceCloningStatus
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

export interface WorkspaceMetrics {
  cpu: {
    usage: number // percentage 0-100
    limit?: string
  }
  memory: {
    usage: number // in bytes
    usagePercent: number // percentage 0-100
    limit?: string
  }
  timestamp: string
}

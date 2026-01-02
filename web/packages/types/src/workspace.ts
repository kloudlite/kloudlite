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

// Visibility is used by environments (not workspaces)
export type Visibility = 'private' | 'shared' | 'open'

export interface EnvironmentConnection {
  environmentRef: ObjectReference
}

export interface WorkspaceSpec {
  displayName: string
  ownedBy: string
  workmachine: string
  environmentConnection?: EnvironmentConnection
  settings?: WorkspaceSettings
  status?: 'active' | 'suspended' | 'archived'
  vscodeVersion?: string
  gitRepository?: GitRepository
  copyFrom?: string
  expose?: ExposedPort[]
}

// ExposedPort - matches Go API (only port, no protocol)
export interface ExposedPort {
  port: number
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

export interface WorkspaceSettings {
  idleTimeout?: number
  startupScript?: string
  environmentVariables?: Record<string, string>
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

export interface WorkspaceCloningStatus {
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

export interface WorkspaceSourceCloningStatus {
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
  // Package status is now read from PackageRequest resource directly
  // (not synced to Workspace.status anymore)
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
  // Idle detection fields
  idleState?: 'active' | 'idle' | string
  idleSince?: string // ISO timestamp of when idle started
  lastActivityTime?: string // ISO timestamp of last activity
  cloningStatus?: WorkspaceCloningStatus
  sourceCloningStatus?: WorkspaceSourceCloningStatus
  // Hash and subdomain for VPN-accessible URLs
  hash?: string // 8-character hash derived from owner-workspaceName
  subdomain?: string // Subdomain from workmachine (e.g., "beanbag.khost.dev")
  // Exposed HTTP routes - keys are port numbers, values are full URLs
  exposedRoutes?: Record<string, string> // e.g., {"3000": "https://p3000-a1b2c3d4.example.khost.dev"}
  // Last restored snapshot tracking
  lastRestoredSnapshot?: {
    name: string
    restoredAt: string
  }
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

// PackageRequest is the source of truth for package installation status
// It's a cluster-scoped resource owned by the Workspace
export interface PackageRequest {
  metadata: {
    name: string
    creationTimestamp?: string
    resourceVersion?: string
    uid?: string
    generation?: number
  }
  spec: {
    workspaceRef: string
    packages: PackageSpec[]
    profileName: string
  }
  status?: {
    observedGeneration?: number
    phase?: 'Pending' | 'Installing' | 'Ready' | 'Failed'
    message?: string
    // New simplified status fields
    packages?: string[] // Just package names for display
    packageCount?: number
    failedPackage?: string // Single package that caused failure
    packagesPath?: string // e.g., /nix/profiles/kloudlite/<workspace>/packages
    profileStorePath?: string // Nix store path of the built environment
    specHash?: string // For change detection
    lastUpdated?: string
  }
}

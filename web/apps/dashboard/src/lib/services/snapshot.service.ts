import { apiClient } from '../api-client'

// Snapshot as returned by the API (flat structure)
export interface Snapshot {
  name: string
  namespace?: string // Namespace where the snapshot lives
  description?: string
  state: 'Pending' | 'Creating' | 'Ready' | 'Uploading' | 'Restoring' | 'Deleting' | 'Pushing' | 'Pulling' | 'Failed' | 'Completed' | ''
  sizeHuman?: string
  sizeBytes?: number
  createdAt?: string
  registry?: {
    endpoint?: string
    repository?: string
    tag?: string
    digest?: string
  }
  parent?: string
  refCount?: number
  message?: string
}

export interface SnapshotListResponse {
  snapshots: Snapshot[]
  count: number
}

export interface CreateSnapshotRequest {
  description?: string
  includeMetadata?: boolean
  keepForDays?: number
}

export interface CreateSnapshotResponse {
  message: string
  snapshot?: Snapshot
  request?: {
    name: string
    snapshotName: string
    phase: string
    message?: string
  }
}

export interface RestoreSnapshotResponse {
  message: string
  snapshot: Snapshot
}

export interface PushSnapshotRequest {
  tag: string
  repository?: string
}

export interface PushSnapshotResponse {
  message: string
  snapshot: string
  tag: string
}

export interface PullSnapshotRequest {
  repository: string
  tag: string
  name?: string
}

export interface PullSnapshotResponse {
  message: string
  snapshot: Snapshot
}

export interface PushedSnapshotListResponse {
  snapshots: Snapshot[]
  count: number
}

export interface CreateWorkspaceFromSnapshotRequest {
  name: string
  displayName?: string
  snapshotName: string
}

export interface CreateEnvironmentFromSnapshotRequest {
  name: string
  snapshotName: string
  sourceNamespace: string // Required: namespace where the source snapshot lives
  activated?: boolean
}

export interface RestoreEnvironmentFromSnapshotRequest {
  snapshotName: string
  sourceNamespace?: string // Optional: defaults to environment's target namespace. Provide for cross-namespace restores (forking)
  activateAfterRestore?: boolean
}

export interface RestoreEnvironmentFromSnapshotResponse {
  message: string
  restore: {
    name: string
    snapshotName: string
    phase: string
    message?: string
  }
}

export interface SnapshotOperationStatus {
  inProgress: boolean
  operation?: 'creating' | 'restoring'
  name?: string
  phase?: string
  message?: string
  snapshotName?: string
}

export interface CreateFromSnapshotResponse {
  message: string
  workspace?: unknown
  environment?: unknown
}

// Fork status response - indicates whether an environment can be forked
export interface ForkStatusResponse {
  canFork: boolean
  latestSnapshot?: string
  namespace?: string
  message?: string
}

// Fork environment request - just needs the new environment name
export interface ForkEnvironmentRequest {
  name: string
}

// Fork environment response
export interface ForkEnvironmentResponse {
  message: string
  forkRequest: string
  sourceEnvironment: string
  newEnvironment: string
  snapshot: string
  phase: string
}

export class SnapshotService {
  private baseUrl = '/api/v1'

  // List snapshots for a workspace
  async listWorkspaceSnapshots(
    workspaceName: string,
    namespace: string,
  ): Promise<SnapshotListResponse> {
    return apiClient.get<SnapshotListResponse>(
      `${this.baseUrl}/namespaces/${namespace}/workspaces/${workspaceName}/snapshots`,
    )
  }

  // Create a snapshot for a workspace
  async createWorkspaceSnapshot(
    workspaceName: string,
    namespace: string,
    data?: CreateSnapshotRequest,
  ): Promise<CreateSnapshotResponse> {
    return apiClient.post<CreateSnapshotResponse>(
      `${this.baseUrl}/namespaces/${namespace}/workspaces/${workspaceName}/snapshots`,
      data || {},
    )
  }

  // List snapshots for an environment
  async listEnvironmentSnapshots(
    environmentName: string,
  ): Promise<SnapshotListResponse> {
    return apiClient.get<SnapshotListResponse>(
      `${this.baseUrl}/environments/${environmentName}/snapshots`,
    )
  }

  // Create a snapshot for an environment
  async createEnvironmentSnapshot(
    environmentName: string,
    data?: CreateSnapshotRequest,
  ): Promise<CreateSnapshotResponse> {
    return apiClient.post<CreateSnapshotResponse>(
      `${this.baseUrl}/environments/${environmentName}/snapshots`,
      data || {},
    )
  }

  // Get a specific snapshot
  async get(snapshotName: string): Promise<Snapshot> {
    return apiClient.get<Snapshot>(`${this.baseUrl}/snapshots/${snapshotName}`)
  }

  // Restore a snapshot
  async restore(snapshotName: string): Promise<RestoreSnapshotResponse> {
    return apiClient.post<RestoreSnapshotResponse>(
      `${this.baseUrl}/snapshots/${snapshotName}/restore`,
    )
  }

  // Delete a snapshot
  async delete(snapshotName: string, namespace: string): Promise<void> {
    return apiClient.delete<void>(`${this.baseUrl}/snapshots/${snapshotName}?namespace=${namespace}`)
  }

  // Push a snapshot to the registry
  async push(
    snapshotName: string,
    tag: string,
    repository?: string,
  ): Promise<PushSnapshotResponse> {
    return apiClient.post<PushSnapshotResponse>(
      `${this.baseUrl}/snapshots/${snapshotName}/push`,
      { tag, repository },
    )
  }

  // Pull a snapshot from the registry
  async pull(
    repository: string,
    tag: string,
    name?: string,
  ): Promise<PullSnapshotResponse> {
    return apiClient.post<PullSnapshotResponse>(
      `${this.baseUrl}/snapshots/pull`,
      { repository, tag, name },
    )
  }

  // Get existing tags for a snapshot repository
  async getTags(repository: string): Promise<string[]> {
    try {
      const response = await apiClient.get<{ name: string; tags: string[] }>(
        `${this.baseUrl}/registry/repositories/${encodeURIComponent(repository)}/tags`,
      )
      return response.tags || []
    } catch {
      return []
    }
  }

  // List ready snapshots available for forking
  // type: filter by snapshot type (workspace or environment)
  // environment: filter by specific environment name
  async listReady(type?: 'workspace' | 'environment', environment?: string): Promise<PushedSnapshotListResponse> {
    const queryParams = new URLSearchParams()
    if (type) queryParams.set('type', type)
    if (environment) queryParams.set('environment', environment)
    const params = queryParams.toString() ? `?${queryParams.toString()}` : ''
    return apiClient.get<PushedSnapshotListResponse>(
      `${this.baseUrl}/snapshots/ready${params}`,
    )
  }

  // Create a workspace from a pushed snapshot
  async createWorkspaceFromSnapshot(
    data: CreateWorkspaceFromSnapshotRequest,
  ): Promise<CreateFromSnapshotResponse> {
    return apiClient.post<CreateFromSnapshotResponse>(
      `${this.baseUrl}/workspaces/from-snapshot`,
      data,
    )
  }

  // Create an environment from a pushed snapshot
  async createEnvironmentFromSnapshot(
    data: CreateEnvironmentFromSnapshotRequest,
  ): Promise<CreateFromSnapshotResponse> {
    return apiClient.post<CreateFromSnapshotResponse>(
      `${this.baseUrl}/environments/from-snapshot`,
      data,
    )
  }

  // Restore an existing environment from a snapshot
  // This stops workloads, restores data, applies artifacts, and optionally activates the environment
  async restoreEnvironmentFromSnapshot(
    environmentName: string,
    data: RestoreEnvironmentFromSnapshotRequest,
  ): Promise<RestoreEnvironmentFromSnapshotResponse> {
    return apiClient.post<RestoreEnvironmentFromSnapshotResponse>(
      `${this.baseUrl}/environments/${environmentName}/restore`,
      data,
    )
  }

  // Get the current snapshot operation status for an environment
  // Returns whether a snapshot creation or restore is in progress
  async getEnvironmentSnapshotStatus(
    environmentName: string,
  ): Promise<SnapshotOperationStatus> {
    return apiClient.get<SnapshotOperationStatus>(
      `${this.baseUrl}/environments/${environmentName}/snapshots/status`,
    )
  }

  // Get fork status for an environment
  // Returns whether the environment can be forked (has ready snapshots)
  async getForkStatus(environmentName: string): Promise<ForkStatusResponse> {
    return apiClient.get<ForkStatusResponse>(
      `${this.baseUrl}/environments/${environmentName}/fork-status`,
    )
  }

  // Fork an environment using the latest ready snapshot
  // Creates a new environment from the source environment's latest snapshot
  async forkEnvironment(
    sourceEnvironmentName: string,
    data: ForkEnvironmentRequest,
  ): Promise<ForkEnvironmentResponse> {
    return apiClient.post<ForkEnvironmentResponse>(
      `${this.baseUrl}/environments/${sourceEnvironmentName}/fork`,
      data,
    )
  }
}

// Export singleton instance
export const snapshotService = new SnapshotService()

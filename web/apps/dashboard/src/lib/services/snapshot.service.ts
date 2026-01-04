import { apiClient } from '../api-client'

export interface Snapshot {
  metadata: {
    name: string
    creationTimestamp: string
    labels?: Record<string, string>
  }
  spec: {
    workspaceRef?: {
      name: string
      workmachineName: string
    }
    environmentRef?: {
      name: string
    }
    parentSnapshotRef?: {
      name: string
      restoredAt?: string
    }
    description?: string
    ownedBy: string
    includeMetadata: boolean
    retentionPolicy?: {
      keepForDays?: number
      expiresAt?: string
    }
  }
  status: {
    state:
      | 'Pending'
      | 'Creating'
      | 'Ready'
      | 'Restoring'
      | 'Deleting'
      | 'Pushing'
      | 'Pulling'
      | 'Failed'
    snapshotType?: 'Workspace' | 'Environment'
    targetName?: string
    message?: string
    sizeBytes?: number
    sizeHuman?: string
    createdAt?: string
    snapshotPath?: string
    workMachineName?: string
    // Registry status for pushed snapshots
    registryStatus?: {
      pushed: boolean
      pushedAt?: string
      tag?: string
      imageRef?: string
      digest?: string
      layerDigests?: string[]
      layerCount?: number
      compressedSize?: number
    }
  }
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
  snapshot: Snapshot
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
  targetNamespace?: string
  activated?: boolean
}

export interface CreateFromSnapshotResponse {
  message: string
  workspace?: unknown
  environment?: unknown
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
  async delete(snapshotName: string): Promise<void> {
    return apiClient.delete<void>(`${this.baseUrl}/snapshots/${snapshotName}`)
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

  // List pushed snapshots available for cloning
  async listPushed(type?: 'workspace' | 'environment'): Promise<PushedSnapshotListResponse> {
    const params = type ? `?type=${type}` : ''
    return apiClient.get<PushedSnapshotListResponse>(
      `${this.baseUrl}/snapshots/pushed${params}`,
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
}

// Export singleton instance
export const snapshotService = new SnapshotService()

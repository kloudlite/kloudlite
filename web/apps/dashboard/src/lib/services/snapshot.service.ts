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
    description?: string
    ownedBy: string
    includeMetadata: boolean
    retentionPolicy?: {
      keepForDays?: number
      expiresAt?: string
    }
  }
  status: {
    state: 'Pending' | 'Creating' | 'Ready' | 'Restoring' | 'Deleting' | 'Failed'
    snapshotType?: 'Workspace' | 'Environment'
    targetName?: string
    message?: string
    sizeBytes?: number
    sizeHuman?: string
    createdAt?: string
    snapshotPath?: string
    workMachineName?: string
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
}

// Export singleton instance
export const snapshotService = new SnapshotService()

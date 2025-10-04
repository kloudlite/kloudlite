import { apiClient } from '../api-client'
import type {
  Workspace,
  WorkspaceCreateRequest,
  WorkspaceUpdateRequest,
  WorkspaceListResponse,
  WorkspaceListParams,
  WorkspaceActionResponse,
} from '@/types/workspace'

export class WorkspaceService {
  private baseUrl = '/api/v1'

  // List workspaces in a namespace
  async list(namespace: string = 'default', params?: WorkspaceListParams): Promise<WorkspaceListResponse> {
    const queryParams = new URLSearchParams()
    if (params?.owner) queryParams.append('owner', params.owner)
    if (params?.workMachine) queryParams.append('workMachine', params.workMachine)
    if (params?.status) queryParams.append('status', params.status)
    if (params?.limit) queryParams.append('limit', params.limit.toString())
    if (params?.continue) queryParams.append('continue', params.continue)

    const query = queryParams.toString()
    const url = `${this.baseUrl}/namespaces/${namespace}/workspaces${query ? `?${query}` : ''}`
    return apiClient.get<WorkspaceListResponse>(url)
  }

  // Get a specific workspace
  async get(name: string, namespace: string = 'default'): Promise<Workspace> {
    return apiClient.get<Workspace>(`${this.baseUrl}/namespaces/${namespace}/workspaces/${name}`)
  }

  // Create a new workspace
  async create(data: WorkspaceCreateRequest, namespace: string = 'default'): Promise<Workspace> {
    return apiClient.post<Workspace>(`${this.baseUrl}/namespaces/${namespace}/workspaces`, data)
  }

  // Update an existing workspace
  async update(name: string, data: WorkspaceUpdateRequest, namespace: string = 'default'): Promise<Workspace> {
    return apiClient.put<Workspace>(`${this.baseUrl}/namespaces/${namespace}/workspaces/${name}`, data)
  }

  // Delete a workspace
  async delete(name: string, namespace: string = 'default'): Promise<void> {
    return apiClient.delete<void>(`${this.baseUrl}/namespaces/${namespace}/workspaces/${name}`)
  }

  // Workspace actions
  async suspend(name: string, namespace: string = 'default'): Promise<WorkspaceActionResponse> {
    return apiClient.post<WorkspaceActionResponse>(
      `${this.baseUrl}/namespaces/${namespace}/workspaces/${name}/suspend`
    )
  }

  async activate(name: string, namespace: string = 'default'): Promise<WorkspaceActionResponse> {
    return apiClient.post<WorkspaceActionResponse>(
      `${this.baseUrl}/namespaces/${namespace}/workspaces/${name}/activate`
    )
  }

  async archive(name: string, namespace: string = 'default'): Promise<WorkspaceActionResponse> {
    return apiClient.post<WorkspaceActionResponse>(
      `${this.baseUrl}/namespaces/${namespace}/workspaces/${name}/archive`
    )
  }
}

// Export singleton instance
export const workspaceService = new WorkspaceService()
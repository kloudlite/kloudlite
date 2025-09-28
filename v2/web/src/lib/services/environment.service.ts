import { apiClient } from '@/lib/api-client'
import type {
  Environment,
  EnvironmentCreateRequest,
  EnvironmentUpdateRequest,
  EnvironmentListResponse,
  EnvironmentResponse,
  EnvironmentDeleteResponse,
  EnvironmentStatusResponse,
} from '@/types/environment'

export class EnvironmentService {
  private baseUrl = '/api/v1/environments'

  /**
   * List all environments
   */
  async listEnvironments(user?: string): Promise<EnvironmentListResponse> {
    const headers: Record<string, string> = {}
    if (user) {
      headers['X-User'] = user
    }

    return apiClient.get<EnvironmentListResponse>(this.baseUrl, { headers })
  }

  /**
   * Get a specific environment by name
   */
  async getEnvironment(name: string, user?: string): Promise<Environment> {
    const headers: Record<string, string> = {}
    if (user) {
      headers['X-User'] = user
    }

    const response = await apiClient.get<Environment>(`${this.baseUrl}/${name}`, { headers })
    return response
  }

  /**
   * Create a new environment
   */
  async createEnvironment(data: EnvironmentCreateRequest, user?: string): Promise<EnvironmentResponse> {
    const headers: Record<string, string> = {}
    if (user) {
      headers['X-User'] = user
    }

    return apiClient.post<EnvironmentResponse>(this.baseUrl, data, { headers })
  }

  /**
   * Update an environment
   */
  async updateEnvironment(
    name: string,
    data: EnvironmentUpdateRequest,
    user?: string
  ): Promise<EnvironmentResponse> {
    const headers: Record<string, string> = {}
    if (user) {
      headers['X-User'] = user
    }

    return apiClient.put<EnvironmentResponse>(`${this.baseUrl}/${name}`, data, { headers })
  }

  /**
   * Delete an environment
   */
  async deleteEnvironment(name: string, user?: string): Promise<EnvironmentDeleteResponse> {
    const headers: Record<string, string> = {}
    if (user) {
      headers['X-User'] = user
    }

    return apiClient.delete<EnvironmentDeleteResponse>(`${this.baseUrl}/${name}`, { headers })
  }

  /**
   * Activate an environment
   */
  async activateEnvironment(name: string, user?: string): Promise<EnvironmentResponse> {
    const headers: Record<string, string> = {}
    if (user) {
      headers['X-User'] = user
    }

    return apiClient.post<EnvironmentResponse>(
      `${this.baseUrl}/${name}/activate`,
      undefined,
      { headers }
    )
  }

  /**
   * Deactivate an environment
   */
  async deactivateEnvironment(name: string, user?: string): Promise<EnvironmentResponse> {
    const headers: Record<string, string> = {}
    if (user) {
      headers['X-User'] = user
    }

    return apiClient.post<EnvironmentResponse>(
      `${this.baseUrl}/${name}/deactivate`,
      undefined,
      { headers }
    )
  }

  /**
   * Get environment status
   */
  async getEnvironmentStatus(name: string, user?: string): Promise<EnvironmentStatusResponse> {
    const headers: Record<string, string> = {}
    if (user) {
      headers['X-User'] = user
    }

    return apiClient.get<EnvironmentStatusResponse>(`${this.baseUrl}/${name}/status`, { headers })
  }
}

// Export singleton instance
export const environmentService = new EnvironmentService()
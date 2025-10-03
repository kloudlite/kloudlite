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
  async listEnvironments(): Promise<EnvironmentListResponse> {
    return apiClient.get<EnvironmentListResponse>(this.baseUrl)
  }

  /**
   * Get a specific environment by name
   */
  async getEnvironment(name: string): Promise<Environment> {
    const response = await apiClient.get<Environment>(`${this.baseUrl}/${name}`)
    return response
  }

  /**
   * Create a new environment
   */
  async createEnvironment(data: EnvironmentCreateRequest): Promise<EnvironmentResponse> {
    return apiClient.post<EnvironmentResponse>(this.baseUrl, data)
  }

  /**
   * Update an environment
   */
  async updateEnvironment(
    name: string,
    data: EnvironmentUpdateRequest
  ): Promise<EnvironmentResponse> {
    return apiClient.put<EnvironmentResponse>(`${this.baseUrl}/${name}`, data)
  }

  /**
   * Delete an environment
   */
  async deleteEnvironment(name: string): Promise<EnvironmentDeleteResponse> {
    return apiClient.delete<EnvironmentDeleteResponse>(`${this.baseUrl}/${name}`)
  }

  /**
   * Activate an environment
   */
  async activateEnvironment(name: string): Promise<EnvironmentResponse> {
    return apiClient.post<EnvironmentResponse>(
      `${this.baseUrl}/${name}/activate`,
      undefined
    )
  }

  /**
   * Deactivate an environment
   */
  async deactivateEnvironment(name: string): Promise<EnvironmentResponse> {
    return apiClient.post<EnvironmentResponse>(
      `${this.baseUrl}/${name}/deactivate`,
      undefined
    )
  }

  /**
   * Get environment status
   */
  async getEnvironmentStatus(name: string): Promise<EnvironmentStatusResponse> {
    return apiClient.get<EnvironmentStatusResponse>(`${this.baseUrl}/${name}/status`)
  }
}

// Export singleton instance
export const environmentService = new EnvironmentService()

import { apiClient } from '@/lib/api-client'
import type {
  MachineType,
  MachineTypeCreateRequest,
  MachineTypeUpdateRequest,
  MachineTypeListResponse,
  MachineTypeResponse,
  MachineTypeDeleteResponse
} from '@/types/machine'

export class MachineTypeService {
  private baseUrl = '/api/v1/machine-types'

  /**
   * List all machine types
   */
  async listMachineTypes(user?: string): Promise<MachineTypeListResponse> {
    const headers: Record<string, string> = {}
    if (user) {
      headers['X-User-Email'] = user
    }

    return apiClient.get<MachineTypeListResponse>(this.baseUrl, { headers })
  }

  /**
   * Get a specific machine type by name
   */
  async getMachineType(name: string, user?: string): Promise<MachineType> {
    const headers: Record<string, string> = {}
    if (user) {
      headers['X-User-Email'] = user
    }

    return apiClient.get<MachineType>(`${this.baseUrl}/${name}`, { headers })
  }

  /**
   * Create a new machine type
   */
  async createMachineType(data: MachineTypeCreateRequest, user?: string): Promise<MachineTypeResponse> {
    const headers: Record<string, string> = {}
    if (user) {
      headers['X-User-Email'] = user
    }

    return apiClient.post<MachineTypeResponse>(this.baseUrl, data, { headers })
  }

  /**
   * Update an existing machine type
   */
  async updateMachineType(name: string, data: MachineTypeUpdateRequest, user?: string): Promise<MachineTypeResponse> {
    const headers: Record<string, string> = {}
    if (user) {
      headers['X-User-Email'] = user
    }

    return apiClient.put<MachineTypeResponse>(`${this.baseUrl}/${name}`, data, { headers })
  }

  /**
   * Delete a machine type
   */
  async deleteMachineType(name: string, user?: string): Promise<MachineTypeDeleteResponse> {
    const headers: Record<string, string> = {}
    if (user) {
      headers['X-User-Email'] = user
    }

    return apiClient.delete<MachineTypeDeleteResponse>(`${this.baseUrl}/${name}`, { headers })
  }

  /**
   * Activate a machine type
   */
  async activateMachineType(name: string, user?: string): Promise<MachineTypeResponse> {
    const headers: Record<string, string> = {}
    if (user) {
      headers['X-User-Email'] = user
    }

    return apiClient.put<MachineTypeResponse>(`${this.baseUrl}/${name}/activate`, {}, { headers })
  }

  /**
   * Deactivate a machine type
   */
  async deactivateMachineType(name: string, user?: string): Promise<MachineTypeResponse> {
    const headers: Record<string, string> = {}
    if (user) {
      headers['X-User-Email'] = user
    }

    return apiClient.put<MachineTypeResponse>(`${this.baseUrl}/${name}/deactivate`, {}, { headers })
  }
}

// Export singleton instance
export const machineTypeService = new MachineTypeService()
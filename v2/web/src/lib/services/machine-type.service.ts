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
  async listMachineTypes(): Promise<MachineTypeListResponse> {
    return apiClient.get<MachineTypeListResponse>(this.baseUrl)
  }

  /**
   * Get a specific machine type by name
   */
  async getMachineType(name: string): Promise<MachineType> {
    return apiClient.get<MachineType>(`${this.baseUrl}/${name}`)
  }

  /**
   * Create a new machine type
   */
  async createMachineType(data: MachineTypeCreateRequest): Promise<MachineTypeResponse> {
    // Transform frontend data structure to backend API format
    const payload = {
      name: data.name,
      spec: {
        displayName: data.displayName || data.name,
        description: data.description || '',
        resources: {
          cpu: data.cpu.toString(),
          memory: `${data.memory}Gi`,
          ...(data.gpu && { gpu: data.gpu.toString() })
        },
        category: data.category,
        active: data.active !== undefined ? data.active : true,
        isDefault: false
      }
    }
    return apiClient.post<MachineTypeResponse>(this.baseUrl, payload)
  }

  /**
   * Update an existing machine type
   */
  async updateMachineType(name: string, data: MachineTypeUpdateRequest): Promise<MachineTypeResponse> {
    // Transform frontend data structure to backend API format
    const payload = {
      spec: {
        ...(data.displayName && { displayName: data.displayName }),
        ...(data.description !== undefined && { description: data.description }),
        ...(data.cpu && {
          resources: {
            cpu: data.cpu.toString(),
            memory: data.memory ? `${data.memory}Gi` : undefined,
            ...(data.gpu && { gpu: data.gpu.toString() })
          }
        }),
        ...(data.category && { category: data.category }),
        ...(data.active !== undefined && { active: data.active })
      }
    }
    return apiClient.put<MachineTypeResponse>(`${this.baseUrl}/${name}`, payload)
  }

  /**
   * Delete a machine type
   */
  async deleteMachineType(name: string): Promise<MachineTypeDeleteResponse> {
    return apiClient.delete<MachineTypeDeleteResponse>(`${this.baseUrl}/${name}`)
  }

  /**
   * Activate a machine type
   */
  async activateMachineType(name: string): Promise<MachineTypeResponse> {
    return apiClient.put<MachineTypeResponse>(`${this.baseUrl}/${name}/activate`, {})
  }

  /**
   * Deactivate a machine type
   */
  async deactivateMachineType(name: string): Promise<MachineTypeResponse> {
    return apiClient.put<MachineTypeResponse>(`${this.baseUrl}/${name}/deactivate`, {})
  }

  /**
   * Set a machine type as default
   */
  async setMachineTypeAsDefault(name: string): Promise<MachineTypeResponse> {
    return apiClient.put<MachineTypeResponse>(`${this.baseUrl}/${name}/set-default`, {})
  }
}

// Export singleton instance
export const machineTypeService = new MachineTypeService()
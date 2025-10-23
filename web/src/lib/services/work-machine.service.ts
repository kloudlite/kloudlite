import { apiClient } from '@/lib/api-client'
import type {
  WorkMachine,
  WorkMachineListResponse,
} from '@/types/work-machine'

export class WorkMachineService {
  private baseUrl = '/api/v1/work-machines'

  /**
   * Get current user's work machine
   */
  async getMyWorkMachine(): Promise<WorkMachine | null> {
    try {
      return await apiClient.get<WorkMachine>(`${this.baseUrl}/my`)
    } catch (err) {
      const error = err instanceof Error ? err : new Error('Unknown error')
      if (error.message?.includes('404')) {
        return null
      }
      throw error
    }
  }

  /**
   * List all work machines (admin only)
   */
  async listAllWorkMachines(): Promise<WorkMachineListResponse> {
    return apiClient.get<WorkMachineListResponse>(this.baseUrl)
  }

  /**
   * Start current user's work machine
   */
  async startMyWorkMachine(): Promise<{message: string; state: string}> {
    return apiClient.post<{message: string; state: string}>(`${this.baseUrl}/my/start`, undefined)
  }

  /**
   * Stop current user's work machine
   */
  async stopMyWorkMachine(): Promise<{message: string; state: string}> {
    return apiClient.post<{message: string; state: string}>(`${this.baseUrl}/my/stop`, undefined)
  }

  /**
   * Update current user's work machine
   */
  async updateMyWorkMachine(data: { machineType?: string; sshPublicKeys?: string[] }): Promise<WorkMachine> {
    return apiClient.put<WorkMachine>(`${this.baseUrl}/my`, data)
  }
}

// Export singleton instance
export const workMachineService = new WorkMachineService()

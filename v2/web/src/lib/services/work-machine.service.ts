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
  async getMyWorkMachine(user?: string): Promise<WorkMachine | null> {
    const headers: Record<string, string> = {}
    if (user) {
      headers['X-User-Email'] = user
    }

    try {
      return await apiClient.get<WorkMachine>(`${this.baseUrl}/my`, { headers })
    } catch (error: any) {
      if (error.message?.includes('404')) {
        return null
      }
      throw error
    }
  }

  /**
   * List all work machines (admin only)
   */
  async listAllWorkMachines(user?: string): Promise<WorkMachineListResponse> {
    const headers: Record<string, string> = {}
    if (user) {
      headers['X-User-Email'] = user
    }

    return apiClient.get<WorkMachineListResponse>(this.baseUrl, { headers })
  }

  /**
   * Start current user's work machine
   */
  async startMyWorkMachine(user?: string): Promise<{message: string; state: string}> {
    const headers: Record<string, string> = {}
    if (user) {
      headers['X-User-Email'] = user
    }

    return apiClient.post<{message: string; state: string}>(`${this.baseUrl}/my/start`, undefined, { headers })
  }

  /**
   * Stop current user's work machine
   */
  async stopMyWorkMachine(user?: string): Promise<{message: string; state: string}> {
    const headers: Record<string, string> = {}
    if (user) {
      headers['X-User-Email'] = user
    }

    return apiClient.post<{message: string; state: string}>(`${this.baseUrl}/my/stop`, undefined, { headers })
  }

  /**
   * Update current user's work machine
   */
  async updateMyWorkMachine(data: { machineType?: string; sshPublicKeys?: string[] }, user?: string): Promise<WorkMachine> {
    const headers: Record<string, string> = {}
    if (user) {
      headers['X-User-Email'] = user
    }

    return apiClient.put<WorkMachine>(`${this.baseUrl}/my`, data, { headers })
  }
}

// Export singleton instance
export const workMachineService = new WorkMachineService()

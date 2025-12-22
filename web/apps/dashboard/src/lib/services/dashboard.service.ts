import { cache } from 'react'
import { apiClient } from '../api-client'
import type {
  MachineType,
  WorkMachine,
  Workspace,
  Environment,
  UserPreferences,
} from '@kloudlite/types'

// Response type from the dashboard endpoint
export interface DashboardData {
  machineTypes: MachineType[]
  workMachines: WorkMachine[]
  preferences: UserPreferences | null
  pinnedWorkspaces: Workspace[]
  pinnedEnvironments: Environment[]
  isAdmin: boolean
}

export class DashboardService {
  private baseUrl = '/api/v1'

  // Get all dashboard data in a single request
  async getDashboard(): Promise<DashboardData> {
    return apiClient.get<DashboardData>(`${this.baseUrl}/dashboard`)
  }
}

// Export singleton instance
export const dashboardService = new DashboardService()

// Cached version for request deduplication in Server Components
export const getDashboardData = cache(async (): Promise<DashboardData> => {
  return dashboardService.getDashboard()
})

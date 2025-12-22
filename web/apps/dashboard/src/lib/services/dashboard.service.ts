import { cache } from 'react'
import { apiClient } from '../api-client'
import type {
  MachineType,
  WorkMachine,
  Workspace,
  Environment,
  UserPreferences,
  K8sService,
  Composition,
  PackageRequest,
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

// Response type from the environment details endpoint
export interface EnvironmentDetailsData {
  environment: Environment
  services: K8sService[]
  composition: Composition | null
  namespace: string
  envHash: string
  subdomain: string
  isActive: boolean
}

// Response type from the workspaces list endpoint
export interface WorkspacesListData {
  workspaces: Workspace[]
  workMachine: WorkMachine | null
  preferences: UserPreferences | null
  pinnedWorkspaceIds: string[]
  workMachineRunning: boolean
}

// Response type from the environments list endpoint
export interface EnvironmentsListData {
  environments: Environment[]
  workMachine: WorkMachine | null
  preferences: UserPreferences | null
  pinnedEnvironmentIds: string[]
  workMachineRunning: boolean
}

// Response type from the workspace details endpoint
export interface WorkspaceDetailsData {
  workspace: Workspace
  workMachine: WorkMachine | null
  packageRequest: PackageRequest | null
  workMachineRunning: boolean
}

export class DashboardService {
  private baseUrl = '/api/v1'

  // Get all dashboard data in a single request
  async getDashboard(): Promise<DashboardData> {
    return apiClient.get<DashboardData>(`${this.baseUrl}/dashboard`)
  }

  // Get environment details (environment + services + composition)
  async getEnvironmentDetails(name: string): Promise<EnvironmentDetailsData> {
    return apiClient.get<EnvironmentDetailsData>(`${this.baseUrl}/environments/${name}/details`)
  }

  // Get workspaces list with work machine and preferences
  async getWorkspacesListFull(): Promise<WorkspacesListData> {
    return apiClient.get<WorkspacesListData>(`${this.baseUrl}/workspaces/list-full`)
  }

  // Get environments list with work machine and preferences
  async getEnvironmentsListFull(): Promise<EnvironmentsListData> {
    return apiClient.get<EnvironmentsListData>(`${this.baseUrl}/environments/list-full`)
  }

  // Get workspace details (workspace + work machine + packages)
  async getWorkspaceDetails(namespace: string, name: string): Promise<WorkspaceDetailsData> {
    return apiClient.get<WorkspaceDetailsData>(`${this.baseUrl}/namespaces/${namespace}/workspaces/${name}/details`)
  }
}

// Export singleton instance
export const dashboardService = new DashboardService()

// Cached version for request deduplication in Server Components
export const getDashboardData = cache(async (): Promise<DashboardData> => {
  return dashboardService.getDashboard()
})

// Cached version for environment details
export const getEnvironmentDetails = cache(async (name: string): Promise<EnvironmentDetailsData> => {
  return dashboardService.getEnvironmentDetails(name)
})

// Cached version for workspaces list
export const getWorkspacesListFull = cache(async (): Promise<WorkspacesListData> => {
  return dashboardService.getWorkspacesListFull()
})

// Cached version for environments list
export const getEnvironmentsListFull = cache(async (): Promise<EnvironmentsListData> => {
  return dashboardService.getEnvironmentsListFull()
})

// Cached version for workspace details
export const getWorkspaceDetails = cache(async (namespace: string, name: string): Promise<WorkspaceDetailsData> => {
  return dashboardService.getWorkspaceDetails(namespace, name)
})

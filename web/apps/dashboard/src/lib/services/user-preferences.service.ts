import { cache } from 'react'
import { apiClient } from '@/lib/api-client'
import type { UserPreferences, PinWorkspaceRequest, PinEnvironmentRequest } from '@kloudlite/types'

class UserPreferencesServiceImpl {
  private baseUrl = '/api/v1/user-preferences'

  async getMyPreferences(): Promise<UserPreferences> {
    return apiClient.get<UserPreferences>(this.baseUrl)
  }

  async pinWorkspace(data: PinWorkspaceRequest): Promise<void> {
    await apiClient.post(`${this.baseUrl}/pinned-workspaces`, data)
  }

  async unpinWorkspace(data: PinWorkspaceRequest): Promise<void> {
    await apiClient.delete(`${this.baseUrl}/pinned-workspaces`, data)
  }

  async pinEnvironment(data: PinEnvironmentRequest): Promise<void> {
    await apiClient.post(`${this.baseUrl}/pinned-environments`, data)
  }

  async unpinEnvironment(data: PinEnvironmentRequest): Promise<void> {
    await apiClient.delete(`${this.baseUrl}/pinned-environments`, data)
  }
}

export const userPreferencesService = new UserPreferencesServiceImpl()

// Cached version for Server Components
export const getMyPreferences = cache(async (): Promise<UserPreferences> => {
  return userPreferencesService.getMyPreferences()
})

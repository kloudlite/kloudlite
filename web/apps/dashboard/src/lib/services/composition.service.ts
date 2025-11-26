import { apiClient } from '@/lib/api-client'
import type {
  Composition,
  CompositionCreateRequest,
  CompositionUpdateRequest,
  CompositionListResponse,
  CompositionResponse,
  CompositionDeleteResponse,
  CompositionStatusResponse,
} from '@kloudlite/types'

export class CompositionService {
  /**
   * List all compositions in a namespace
   */
  async listCompositions(namespace: string, state?: string): Promise<CompositionListResponse> {
    const baseUrl = `/api/v1/namespaces/${namespace}/compositions`
    const url = state ? `${baseUrl}?state=${state}` : baseUrl

    return apiClient.get<CompositionListResponse>(url)
  }

  /**
   * Get a specific composition by name
   */
  async getComposition(namespace: string, name: string): Promise<Composition> {
    return apiClient.get<Composition>(`/api/v1/namespaces/${namespace}/compositions/${name}`)
  }

  /**
   * Create a new composition
   */
  async createComposition(
    namespace: string,
    data: CompositionCreateRequest,
  ): Promise<CompositionResponse> {
    return apiClient.post<CompositionResponse>(`/api/v1/namespaces/${namespace}/compositions`, data)
  }

  /**
   * Update a composition
   */
  async updateComposition(
    namespace: string,
    name: string,
    data: CompositionUpdateRequest,
  ): Promise<CompositionResponse> {
    return apiClient.put<CompositionResponse>(
      `/api/v1/namespaces/${namespace}/compositions/${name}`,
      data,
    )
  }

  /**
   * Delete a composition
   */
  async deleteComposition(namespace: string, name: string): Promise<CompositionDeleteResponse> {
    return apiClient.delete<CompositionDeleteResponse>(
      `/api/v1/namespaces/${namespace}/compositions/${name}`,
    )
  }

  /**
   * Get composition status
   */
  async getCompositionStatus(namespace: string, name: string): Promise<CompositionStatusResponse> {
    return apiClient.get<CompositionStatusResponse>(
      `/api/v1/namespaces/${namespace}/compositions/${name}/status`,
    )
  }
}

// Export singleton instance
export const compositionService = new CompositionService()

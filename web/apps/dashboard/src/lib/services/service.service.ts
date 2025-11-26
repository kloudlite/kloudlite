import { apiClient } from '@/lib/api-client'
import type { ListServicesResponse } from '@kloudlite/types'

export class ServiceService {
  /**
   * List all services in a namespace (read-only)
   */
  async listServices(namespace: string): Promise<ListServicesResponse> {
    return apiClient.get<ListServicesResponse>(`/api/v1/namespaces/${namespace}/services`)
  }
}

// Export singleton instance
export const serviceService = new ServiceService()

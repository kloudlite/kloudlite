import { apiClient } from '@/lib/api-client'
import type { ListServiceInterceptsResponse, ServiceIntercept } from '@/types/serviceintercept'

export class ServiceInterceptService {
  /**
   * List all service intercepts in a namespace
   */
  async listServiceIntercepts(namespace: string): Promise<ListServiceInterceptsResponse> {
    return apiClient.get<ListServiceInterceptsResponse>(`/api/v1/namespaces/${namespace}/service-intercepts`)
  }

  /**
   * Get a specific service intercept
   */
  async getServiceIntercept(namespace: string, name: string): Promise<ServiceIntercept> {
    return apiClient.get<ServiceIntercept>(`/api/v1/namespaces/${namespace}/service-intercepts/${name}`)
  }

  /**
   * Create a service intercept
   */
  async createServiceIntercept(namespace: string, data: { name: string; spec: ServiceIntercept['spec'] }): Promise<ServiceIntercept> {
    return apiClient.post<ServiceIntercept>(`/api/v1/namespaces/${namespace}/service-intercepts`, data)
  }

  /**
   * Update a service intercept
   */
  async updateServiceIntercept(namespace: string, name: string, spec: ServiceIntercept['spec']): Promise<ServiceIntercept> {
    return apiClient.put<ServiceIntercept>(`/api/v1/namespaces/${namespace}/service-intercepts/${name}`, { spec })
  }

  /**
   * Delete a service intercept
   */
  async deleteServiceIntercept(namespace: string, name: string): Promise<void> {
    return apiClient.delete(`/api/v1/namespaces/${namespace}/service-intercepts/${name}`)
  }

  /**
   * Activate a service intercept
   */
  async activateServiceIntercept(namespace: string, name: string): Promise<ServiceIntercept> {
    return apiClient.post<ServiceIntercept>(`/api/v1/namespaces/${namespace}/service-intercepts/${name}/activate`, {})
  }

  /**
   * Deactivate a service intercept
   */
  async deactivateServiceIntercept(namespace: string, name: string): Promise<ServiceIntercept> {
    return apiClient.post<ServiceIntercept>(`/api/v1/namespaces/${namespace}/service-intercepts/${name}/deactivate`, {})
  }
}

// Export singleton instance
export const serviceInterceptService = new ServiceInterceptService()

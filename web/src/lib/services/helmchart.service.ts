import { apiClient } from '@/lib/api-client'
import type {
  HelmChart,
  HelmChartCreateRequest,
  HelmChartUpdateRequest,
  HelmChartListResponse,
  HelmChartResponse,
  HelmChartDeleteResponse,
  HelmChartStatusResponse,
} from '@/types/helmchart'

export class HelmChartService {
  /**
   * List all helm charts in a namespace
   */
  async listHelmCharts(namespace: string, repo?: string): Promise<HelmChartListResponse> {
    const baseUrl = `/api/v1/namespaces/${namespace}/helmcharts`
    const url = repo ? `${baseUrl}?repo=${encodeURIComponent(repo)}` : baseUrl

    return apiClient.get<HelmChartListResponse>(url)
  }

  /**
   * Get a specific helm chart by name
   */
  async getHelmChart(namespace: string, name: string): Promise<HelmChart> {
    return apiClient.get<HelmChart>(`/api/v1/namespaces/${namespace}/helmcharts/${name}`)
  }

  /**
   * Create a new helm chart
   */
  async createHelmChart(
    namespace: string,
    data: HelmChartCreateRequest
  ): Promise<HelmChartResponse> {
    return apiClient.post<HelmChartResponse>(
      `/api/v1/namespaces/${namespace}/helmcharts`,
      data
    )
  }

  /**
   * Update a helm chart
   */
  async updateHelmChart(
    namespace: string,
    name: string,
    data: HelmChartUpdateRequest
  ): Promise<HelmChartResponse> {
    return apiClient.put<HelmChartResponse>(
      `/api/v1/namespaces/${namespace}/helmcharts/${name}`,
      data
    )
  }

  /**
   * Delete a helm chart
   */
  async deleteHelmChart(
    namespace: string,
    name: string
  ): Promise<HelmChartDeleteResponse> {
    return apiClient.delete<HelmChartDeleteResponse>(
      `/api/v1/namespaces/${namespace}/helmcharts/${name}`
    )
  }

  /**
   * Get helm chart status
   */
  async getHelmChartStatus(
    namespace: string,
    name: string
  ): Promise<HelmChartStatusResponse> {
    return apiClient.get<HelmChartStatusResponse>(
      `/api/v1/namespaces/${namespace}/helmcharts/${name}/status`
    )
  }
}

// Export singleton instance
export const helmChartService = new HelmChartService()

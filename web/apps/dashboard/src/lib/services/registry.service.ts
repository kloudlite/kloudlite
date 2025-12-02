import { apiClient } from '../api-client'

// Repository info from the API
export interface RepositoryInfo {
  name: string
}

// Response from listing repositories
export interface RepositoryListResponse {
  repositories: RepositoryInfo[]
}

// Response from listing tags
export interface TagListResponse {
  name: string
  tags: string[]
}

export class RegistryService {
  private baseUrl = '/api/v1/registry'

  // List all repositories in the registry
  async listRepositories(): Promise<RepositoryListResponse> {
    return apiClient.get<RepositoryListResponse>(`${this.baseUrl}/repositories`)
  }

  // List all tags for a specific repository
  async listTags(repository: string): Promise<TagListResponse> {
    // Repository name may contain slashes (e.g., "karthik/myapp")
    return apiClient.get<TagListResponse>(`${this.baseUrl}/repositories/${repository}`)
  }

  // Delete a specific tag from a repository
  async deleteTag(repository: string, tag: string): Promise<void> {
    const params = new URLSearchParams({ repo: repository, tag })
    await apiClient.delete(`${this.baseUrl}/tags?${params.toString()}`)
  }
}

// Export singleton instance
export const registryService = new RegistryService()

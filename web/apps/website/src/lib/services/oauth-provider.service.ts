import { apiClient } from '@/lib/api-client'

export interface OAuthProvider {
  type: string
  enabled: boolean
  clientId: string
  clientSecret?: string
}

export class OAuthProviderService {
  private baseUrl = '/api/v1/oauth-providers'

  /**
   * Get all OAuth providers
   */
  async getOAuthProviders(): Promise<Record<string, OAuthProvider>> {
    return apiClient.get<Record<string, OAuthProvider>>(this.baseUrl)
  }

  /**
   * Update an OAuth provider
   */
  async updateOAuthProvider(type: string, provider: OAuthProvider): Promise<{ success: boolean }> {
    return apiClient.put<{ success: boolean }>(`${this.baseUrl}/${type}`, provider)
  }
}

// Export singleton instance
export const oauthProviderService = new OAuthProviderService()

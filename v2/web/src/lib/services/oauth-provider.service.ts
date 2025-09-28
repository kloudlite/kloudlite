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
  async getOAuthProviders(user?: string): Promise<Record<string, OAuthProvider>> {
    const headers: Record<string, string> = {}
    if (user) {
      headers['X-User-Email'] = user
    }

    return apiClient.get<Record<string, OAuthProvider>>(this.baseUrl, { headers })
  }

  /**
   * Update an OAuth provider
   */
  async updateOAuthProvider(type: string, provider: OAuthProvider, user?: string): Promise<{ success: boolean }> {
    const headers: Record<string, string> = {}
    if (user) {
      headers['X-User-Email'] = user
    }

    return apiClient.put<{ success: boolean }>(`${this.baseUrl}/${type}`, provider, { headers })
  }
}

// Export singleton instance
export const oauthProviderService = new OAuthProviderService()
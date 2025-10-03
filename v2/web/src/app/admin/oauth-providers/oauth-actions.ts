'use server'

import { oauthProviderService, type OAuthProvider } from '@/lib/services/oauth-provider.service'

export async function getOAuthProviders(): Promise<Record<string, OAuthProvider>> {
  try {
    return await oauthProviderService.getOAuthProviders()
  } catch (error: any) {
    console.error('Error fetching OAuth providers:', error)
    throw new Error(error.message || 'Failed to fetch OAuth providers')
  }
}

export async function updateOAuthProvider(
  type: string,
  provider: OAuthProvider
): Promise<{ success: boolean; error?: string }> {
  try {
    const result = await oauthProviderService.updateOAuthProvider(type, provider)
    return result
  } catch (error: any) {
    console.error('Error updating OAuth provider:', error)
    return { success: false, error: error.message || 'Failed to update OAuth provider' }
  }
}

'use server'

import { oauthProviderService, type OAuthProvider } from '@/lib/services/oauth-provider.service'

export async function getOAuthProviders(): Promise<Record<string, OAuthProvider>> {
  try {
    return await oauthProviderService.getOAuthProviders()
  } catch (err) {
    console.error('Error fetching OAuth providers:', err)
    const error = err instanceof Error ? err : new Error('Failed to fetch OAuth providers')
    throw error
  }
}

export async function updateOAuthProvider(
  type: string,
  provider: OAuthProvider,
): Promise<{ success: boolean; error?: string }> {
  try {
    const result = await oauthProviderService.updateOAuthProvider(type, provider)
    return result
  } catch (err) {
    console.error('Error updating OAuth provider:', err)
    const error = err instanceof Error ? err : new Error('Failed to update OAuth provider')
    return { success: false, error: error.message }
  }
}

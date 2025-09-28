'use server'

import { auth } from '@/lib/auth'
import { oauthProviderService, type OAuthProvider } from '@/lib/services/oauth-provider.service'

export async function getOAuthProviders(): Promise<Record<string, OAuthProvider>> {
  const session = await auth()
  if (!session?.user?.email) {
    throw new Error('Unauthorized')
  }

  try {
    return await oauthProviderService.getOAuthProviders(session.user.email)
  } catch (error: any) {
    console.error('Error fetching OAuth providers:', error)
    throw new Error(error.message || 'Failed to fetch OAuth providers')
  }
}

export async function updateOAuthProvider(
  type: string,
  provider: OAuthProvider
): Promise<{ success: boolean; error?: string }> {
  const session = await auth()
  if (!session?.user?.email) {
    return { success: false, error: 'Unauthorized' }
  }

  try {
    const result = await oauthProviderService.updateOAuthProvider(type, provider, session.user.email)
    return result
  } catch (error: any) {
    console.error('Error updating OAuth provider:', error)
    return { success: false, error: error.message || 'Failed to update OAuth provider' }
  }
}
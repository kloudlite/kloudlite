'use server'

import {
  getOAuthProviders as getProviders,
  updateOAuthProvider as updateProvider,
  type OAuthProvider,
} from '@/lib/services/oauth-provider.service'

export async function getOAuthProviders(): Promise<Record<string, OAuthProvider>> {
  return getProviders()
}

export async function updateOAuthProvider(
  type: string,
  provider: OAuthProvider,
): Promise<{ success: boolean; error?: string }> {
  return updateProvider(type, provider)
}

import {
  getOAuthConfig,
  saveOAuthConfig,
  type OAuthProviderConfig,
} from '@/lib/oauth-config'
import { invalidateAuth } from '@/lib/auth'

export interface OAuthProvider {
  type: string
  enabled: boolean
  clientId: string
  clientSecret?: string
  tenantId?: string
}

/**
 * Get all OAuth providers from the cached config.
 */
export function getOAuthProviders(): Record<string, OAuthProvider> {
  const config = getOAuthConfig()

  return {
    google: {
      type: 'google',
      enabled: config.google.enabled,
      clientId: config.google.clientId,
      clientSecret: config.google.clientSecret,
    },
    github: {
      type: 'github',
      enabled: config.github.enabled,
      clientId: config.github.clientId,
      clientSecret: config.github.clientSecret,
    },
    microsoft: {
      type: 'microsoft',
      enabled: config.microsoft.enabled,
      clientId: config.microsoft.clientId,
      clientSecret: config.microsoft.clientSecret,
      tenantId: config.microsoft.tenantId,
    },
  }
}

/**
 * Update an OAuth provider, save to K8s Secret, and invalidate auth.
 */
export async function updateOAuthProvider(
  type: string,
  data: OAuthProvider
): Promise<{ success: boolean; error?: string }> {
  try {
    const config = getOAuthConfig()

    // Merge the update into the existing config
    const updated: OAuthProviderConfig = { ...config }

    if (type === 'google') {
      updated.google = {
        enabled: data.enabled,
        clientId: data.clientId,
        clientSecret: data.clientSecret || config.google.clientSecret,
      }
    } else if (type === 'github') {
      updated.github = {
        enabled: data.enabled,
        clientId: data.clientId,
        clientSecret: data.clientSecret || config.github.clientSecret,
      }
    } else if (type === 'microsoft') {
      updated.microsoft = {
        enabled: data.enabled,
        clientId: data.clientId,
        clientSecret: data.clientSecret || config.microsoft.clientSecret,
        tenantId: data.tenantId || config.microsoft.tenantId,
      }
    } else {
      return { success: false, error: `Unknown provider type: ${type}` }
    }

    await saveOAuthConfig(updated)
    invalidateAuth()

    return { success: true }
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err)
    console.error(`[OAUTH] Failed to update provider ${type}:`, msg)
    return { success: false, error: msg }
  }
}

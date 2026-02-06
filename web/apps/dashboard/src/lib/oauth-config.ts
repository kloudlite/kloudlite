/**
 * OAuth Provider Configuration via K8s Secret
 *
 * Stores OAuth configs in K8s Secret `kloudlite-oauth-providers` in `default` namespace.
 * No env var fallback — config is exclusively managed via the admin UI.
 * Cache is stored in globalThis to survive Next.js HMR.
 */

const SECRET_NAME = 'kloudlite-oauth-providers'
const SECRET_NAMESPACE = 'default'

export interface OAuthProviderConfig {
  google: { enabled: boolean; clientId: string; clientSecret: string }
  github: { enabled: boolean; clientId: string; clientSecret: string }
  microsoft: { enabled: boolean; clientId: string; clientSecret: string; tenantId: string }
}

const emptyConfig: OAuthProviderConfig = {
  google: { enabled: false, clientId: '', clientSecret: '' },
  github: { enabled: false, clientId: '', clientSecret: '' },
  microsoft: { enabled: false, clientId: '', clientSecret: '', tenantId: '' },
}

// Survive HMR
declare global {
  var __oauthProviderConfig: OAuthProviderConfig | undefined
}

function decodeSecretData(data: Record<string, string>): OAuthProviderConfig {
  const decode = (key: string) => {
    const val = data[key]
    if (!val) return ''
    return Buffer.from(val, 'base64').toString('utf-8')
  }

  return {
    google: {
      enabled: decode('google-enabled') === 'true',
      clientId: decode('google-client-id'),
      clientSecret: decode('google-client-secret'),
    },
    github: {
      enabled: decode('github-enabled') === 'true',
      clientId: decode('github-client-id'),
      clientSecret: decode('github-client-secret'),
    },
    microsoft: {
      enabled: decode('microsoft-enabled') === 'true',
      clientId: decode('microsoft-client-id'),
      clientSecret: decode('microsoft-client-secret'),
      tenantId: decode('microsoft-tenant-id'),
    },
  }
}

function encodeSecretData(config: OAuthProviderConfig): Record<string, string> {
  const encode = (val: string) => Buffer.from(val).toString('base64')

  return {
    'google-enabled': encode(String(config.google.enabled)),
    'google-client-id': encode(config.google.clientId),
    'google-client-secret': encode(config.google.clientSecret),
    'github-enabled': encode(String(config.github.enabled)),
    'github-client-id': encode(config.github.clientId),
    'github-client-secret': encode(config.github.clientSecret),
    'microsoft-enabled': encode(String(config.microsoft.enabled)),
    'microsoft-client-id': encode(config.microsoft.clientId),
    'microsoft-client-secret': encode(config.microsoft.clientSecret),
    'microsoft-tenant-id': encode(config.microsoft.tenantId),
  }
}

/**
 * Load OAuth config from K8s Secret into globalThis cache.
 * Returns empty config (all providers disabled) if secret doesn't exist.
 */
export async function loadOAuthConfig(): Promise<OAuthProviderConfig> {
  try {
    const { secretRepository } = await import('@kloudlite/lib/k8s')
    const secret = await secretRepository.get(SECRET_NAMESPACE, SECRET_NAME)

    if (secret.data) {
      const config = decodeSecretData(secret.data)
      globalThis.__oauthProviderConfig = config
      console.log('[OAUTH] Loaded config from K8s Secret')
      return config
    }
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err)
    if (msg.includes('not found') || msg.includes('404')) {
      console.log('[OAUTH] Secret not found — no OAuth providers configured')
    } else {
      console.warn('[OAUTH] Failed to load secret:', msg)
    }
  }

  globalThis.__oauthProviderConfig = emptyConfig
  return emptyConfig
}

/**
 * Get cached OAuth config (synchronous).
 * Returns empty config if cache isn't populated yet.
 */
export function getOAuthConfig(): OAuthProviderConfig {
  return globalThis.__oauthProviderConfig || emptyConfig
}

/**
 * Save OAuth config to K8s Secret and update cache.
 */
export async function saveOAuthConfig(config: OAuthProviderConfig): Promise<void> {
  const { secretRepository } = await import('@kloudlite/lib/k8s')

  const secretBody = {
    apiVersion: 'v1' as const,
    kind: 'Secret' as const,
    metadata: {
      name: SECRET_NAME,
      namespace: SECRET_NAMESPACE,
    },
    data: encodeSecretData(config),
  }

  try {
    // Try update first
    await secretRepository.update(SECRET_NAMESPACE, SECRET_NAME, secretBody)
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err)
    if (msg.includes('not found') || msg.includes('404')) {
      // Secret doesn't exist yet, create it
      await secretRepository.create(SECRET_NAMESPACE, secretBody)
    } else {
      throw err
    }
  }

  globalThis.__oauthProviderConfig = config
  console.log('[OAUTH] Saved config to K8s Secret')
}

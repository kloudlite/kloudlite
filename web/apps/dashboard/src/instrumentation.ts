/**
 * Next.js Instrumentation
 *
 * This file runs once when the Next.js server starts.
 * Used to initialize server-side dashboard services.
 */

export async function register() {
  // Only run on server side
  if (process.env.NEXT_RUNTIME === 'nodejs') {
    console.log('[INSTRUMENTATION] Initializing server-side services...')

    try {
      const { loadOAuthConfig } = await import('./lib/oauth-config')
      await loadOAuthConfig()
      console.log('[INSTRUMENTATION] OAuth config loaded')
    } catch (err) {
      console.error('[INSTRUMENTATION] Failed to load OAuth config (will use env vars):', err)
      // Non-fatal: auth falls back to env vars
    }
  }
}

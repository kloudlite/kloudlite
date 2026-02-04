/**
 * Next.js Instrumentation
 *
 * This file runs once when the Next.js server starts.
 * Used to initialize K8s watchers for real-time cache invalidation.
 */

export async function register() {
  // Only run on server side
  if (process.env.NEXT_RUNTIME === 'nodejs') {
    console.log('[INSTRUMENTATION] Initializing server-side services...')

    try {
      // Dynamically import to avoid issues with edge runtime
      const { initializeWatchers } = await import('./lib/k8s-watcher')
      await initializeWatchers()
      console.log('[INSTRUMENTATION] K8s watchers initialized')
    } catch (err) {
      console.error('[INSTRUMENTATION] Failed to initialize K8s watchers:', err)
      // Don't throw - allow server to start even if watchers fail
      // The LRU cache will still work with TTL-based expiration
    }
  }
}

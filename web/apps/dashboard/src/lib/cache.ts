import { LRUCache } from 'lru-cache'

/**
 * Server-side LRU cache for K8s API responses
 * Reduces load on API server by caching frequently accessed data
 */

// Only log cache operations in development
const isDev = process.env.NODE_ENV === 'development'

// Cache configuration
const DEFAULT_TTL = 30 * 1000 // 30 seconds
const MAX_ITEMS = 1000

// Create singleton cache instance
const cache = new LRUCache<string, any>({
  max: MAX_ITEMS,
  ttl: DEFAULT_TTL,
  updateAgeOnGet: false, // Don't reset TTL on read
  updateAgeOnHas: false,
})

/**
 * Fetch data with caching
 * @param key - Cache key (should be unique per resource)
 * @param fetcher - Async function to fetch data if not cached
 * @param ttl - Time to live in milliseconds (default: 30s)
 */
export async function cachedFetch<T>(
  key: string,
  fetcher: () => Promise<T>,
  ttl: number = DEFAULT_TTL
): Promise<T> {
  // Check cache first
  const cached = cache.get(key)
  if (cached !== undefined) {
    if (isDev) console.log(`[CACHE HIT] ${key}`)
    return cached as T
  }

  // Cache miss - fetch data
  if (isDev) console.log(`[CACHE MISS] ${key}`)
  const data = await fetcher()

  // Store in cache
  cache.set(key, data, { ttl })
  return data
}

/**
 * Invalidate cache entries by key or pattern
 * @param keyOrPattern - Exact key or pattern to match (prefix match)
 */
export function invalidateCache(keyOrPattern: string): void {
  if (keyOrPattern.endsWith('*')) {
    // Pattern match - invalidate all keys starting with pattern
    const prefix = keyOrPattern.slice(0, -1)
    for (const key of cache.keys()) {
      if (key.startsWith(prefix)) {
        cache.delete(key)
        if (isDev) console.log(`[CACHE INVALIDATE] ${key}`)
      }
    }
  } else {
    // Exact match
    cache.delete(keyOrPattern)
    if (isDev) console.log(`[CACHE INVALIDATE] ${keyOrPattern}`)
  }
}

/**
 * Invalidate all cache entries for a namespace
 */
export function invalidateNamespace(namespace: string): void {
  invalidateCache(`workspaces:${namespace}*`)
  invalidateCache(`environments:${namespace}*`)
}

/**
 * Invalidate all cache entries for a user
 */
export function invalidateUser(username: string): void {
  invalidateCache(`workMachine:${username}*`)
  invalidateCache(`preferences:${username}*`)
}

/**
 * Clear entire cache
 */
export function clearCache(): void {
  cache.clear()
  if (isDev) console.log('[CACHE CLEAR] All entries cleared')
}

/**
 * Get cache statistics
 */
export function getCacheStats() {
  return {
    size: cache.size,
    maxSize: MAX_ITEMS,
    calculatedSize: cache.calculatedSize,
  }
}

// Pre-defined TTLs for different data types
export const CacheTTL = {
  /** Data that rarely changes (5 minutes) - MachineTypes, Snapshots */
  STATIC: 5 * 60 * 1000,
  /** Data that changes occasionally (1 minute) - Users, UserPreferences, Environments */
  NORMAL: 60 * 1000,
  /** Data that changes frequently (30 seconds) - Workspaces, WorkMachines */
  SHORT: 30 * 1000,
  /** Data that changes very frequently (10 seconds) - Metrics, PackageRequests */
  VOLATILE: 10 * 1000,
  /** Infrastructure data (2 minutes) - Services, ConfigMaps, Secrets */
  INFRA: 2 * 60 * 1000,
} as const

// Resource-specific TTL recommendations
export const ResourceTTL = {
  machineTypes: CacheTTL.STATIC,
  workMachine: CacheTTL.SHORT,
  workspaces: CacheTTL.SHORT,
  environments: CacheTTL.NORMAL,
  users: CacheTTL.NORMAL,
  preferences: CacheTTL.NORMAL,
  snapshots: CacheTTL.STATIC,
  packageRequests: CacheTTL.VOLATILE,
  services: CacheTTL.INFRA,
  configmaps: CacheTTL.INFRA,
  secrets: CacheTTL.INFRA,
  metrics: CacheTTL.VOLATILE,
} as const

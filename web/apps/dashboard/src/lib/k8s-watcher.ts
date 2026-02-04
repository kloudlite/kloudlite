import { getK8sClient } from '@kloudlite/lib/k8s'
import { invalidateCache } from './cache'

/**
 * K8s Resource Watcher
 *
 * Watches for changes in K8s resources and invalidates the LRU cache
 * when changes occur. This keeps the cache fresh without polling.
 *
 * Flow:
 * 1. Start watch on K8s resources
 * 2. When resource changes (ADDED/MODIFIED/DELETED), invalidate cache
 * 3. Next read triggers fresh fetch from K8s API
 * 4. LRU cache stores the fresh data
 */

interface WatchConfig {
  group: string
  version: string
  plural: string
  cacheKeyPrefix: string
  namespace?: string
}

const activeWatches = new Map<string, boolean>()
let watchInitialized = false

// Cluster-scoped resources to watch globally
const CLUSTER_WATCH_CONFIGS: WatchConfig[] = [
  {
    group: 'machines.kloudlite.io',
    version: 'v1',
    plural: 'machinetypes',
    cacheKeyPrefix: 'machineTypes',
  },
  {
    group: 'machines.kloudlite.io',
    version: 'v1',
    plural: 'workmachines',
    cacheKeyPrefix: 'workMachine',
  },
  {
    group: 'platform.kloudlite.io',
    version: 'v1alpha1',
    plural: 'users',
    cacheKeyPrefix: 'user',
  },
  {
    group: 'platform.kloudlite.io',
    version: 'v1alpha1',
    plural: 'userpreferences',
    cacheKeyPrefix: 'preferences',
  },
]

// Namespace-scoped resources to watch per namespace
const NAMESPACE_WATCH_CONFIGS: Omit<WatchConfig, 'namespace'>[] = [
  {
    group: 'workspaces.kloudlite.io',
    version: 'v1',
    plural: 'workspaces',
    cacheKeyPrefix: 'workspaces',
  },
  {
    group: 'environments.kloudlite.io',
    version: 'v1',
    plural: 'environments',
    cacheKeyPrefix: 'environments',
  },
  {
    group: 'snapshots.kloudlite.io',
    version: 'v1',
    plural: 'snapshots',
    cacheKeyPrefix: 'snapshots',
  },
  {
    group: 'packages.kloudlite.io',
    version: 'v1',
    plural: 'packagerequests',
    cacheKeyPrefix: 'packageRequests',
  },
  {
    group: '', // core API
    version: 'v1',
    plural: 'services',
    cacheKeyPrefix: 'services',
  },
  {
    group: '', // core API
    version: 'v1',
    plural: 'configmaps',
    cacheKeyPrefix: 'configmaps',
  },
  {
    group: '', // core API
    version: 'v1',
    plural: 'secrets',
    cacheKeyPrefix: 'secrets',
  },
]

/**
 * Start a watch for a specific resource type
 */
async function startWatch(config: WatchConfig): Promise<void> {
  const watchKey = `${config.group || 'core'}/${config.plural}/${config.namespace || 'cluster'}`

  if (activeWatches.get(watchKey)) {
    return // Already watching
  }

  const client = getK8sClient()
  const k8sWatch = client.createWatch()

  // Build the watch path - core API vs custom resources
  let path: string
  if (config.group) {
    // Custom resource (e.g., /apis/workspaces.kloudlite.io/v1/...)
    path = config.namespace
      ? `/apis/${config.group}/${config.version}/namespaces/${config.namespace}/${config.plural}`
      : `/apis/${config.group}/${config.version}/${config.plural}`
  } else {
    // Core API (e.g., /api/v1/...)
    path = config.namespace
      ? `/api/${config.version}/namespaces/${config.namespace}/${config.plural}`
      : `/api/${config.version}/${config.plural}`
  }

  console.log(`[K8S-WATCHER] Starting watch: ${watchKey}`)
  activeWatches.set(watchKey, true)

  try {
    await k8sWatch.watch(
      path,
      {},
      (type: string, obj: any) => {
        const name = obj.metadata?.name
        const namespace = obj.metadata?.namespace
        const owner = obj.spec?.ownedBy

        console.log(`[K8S-WATCHER] ${config.plural} ${type}: ${name}`)

        // Invalidate relevant cache entries based on resource type
        invalidateCacheForResource(config.plural, { name, namespace, owner })
      },
      (err: any) => {
        console.error(`[K8S-WATCHER] Watch error for ${watchKey}:`, err?.message || err)
        activeWatches.set(watchKey, false)

        // Restart watch after delay
        setTimeout(() => {
          console.log(`[K8S-WATCHER] Restarting watch: ${watchKey}`)
          startWatch(config).catch(console.error)
        }, 5000)
      }
    )
  } catch (err) {
    console.error(`[K8S-WATCHER] Failed to start watch ${watchKey}:`, err)
    activeWatches.set(watchKey, false)
  }
}

/**
 * Invalidate cache entries for a specific resource type
 */
function invalidateCacheForResource(
  plural: string,
  meta: { name?: string; namespace?: string; owner?: string }
) {
  const { name, namespace, owner } = meta

  switch (plural) {
    // Cluster-scoped resources
    case 'machinetypes':
      invalidateCache('machineTypes:*')
      break
    case 'workmachines':
      if (owner) {
        invalidateCache(`workMachine:${owner}*`)
      }
      invalidateCache('workMachine:*')
      break
    case 'users':
      if (name) {
        invalidateCache(`user:${name}*`)
      }
      invalidateCache('users:*')
      break
    case 'userpreferences':
      if (name) {
        invalidateCache(`preferences:${name}*`)
      }
      break

    // Namespace-scoped resources
    case 'workspaces':
      if (namespace) {
        invalidateCache(`workspaces:${namespace}*`)
      }
      invalidateCache('workspaces:*')
      break
    case 'environments':
      if (namespace) {
        invalidateCache(`environments:${namespace}*`)
      }
      invalidateCache('environments:*')
      break
    case 'snapshots':
      if (namespace) {
        invalidateCache(`snapshots:${namespace}*`)
      }
      break
    case 'packagerequests':
      if (namespace) {
        invalidateCache(`packageRequests:${namespace}*`)
      }
      break
    case 'services':
      if (namespace) {
        invalidateCache(`services:${namespace}*`)
      }
      break
    case 'configmaps':
      if (namespace) {
        invalidateCache(`configmaps:${namespace}*`)
      }
      break
    case 'secrets':
      if (namespace) {
        invalidateCache(`secrets:${namespace}*`)
      }
      break
  }
}

/**
 * Initialize watches for cluster-scoped resources
 * Call this once when the server starts
 */
export async function initializeWatchers(): Promise<void> {
  if (watchInitialized) {
    return
  }

  console.log('[K8S-WATCHER] Initializing watchers...')
  watchInitialized = true

  // Start watches for cluster-scoped resources
  for (const config of CLUSTER_WATCH_CONFIGS) {
    startWatch(config).catch(console.error)
  }
}

/**
 * Add watches for all resources in a specific namespace
 * Call this when a user accesses a namespace
 */
export async function watchNamespace(namespace: string): Promise<void> {
  console.log(`[K8S-WATCHER] Starting namespace watches for: ${namespace}`)

  for (const config of NAMESPACE_WATCH_CONFIGS) {
    startWatch({
      ...config,
      namespace,
    }).catch(console.error)
  }
}

/**
 * Add a watch for workspaces in a specific namespace (convenience function)
 */
export async function watchNamespaceWorkspaces(namespace: string): Promise<void> {
  await startWatch({
    group: 'workspaces.kloudlite.io',
    version: 'v1',
    plural: 'workspaces',
    cacheKeyPrefix: 'workspaces',
    namespace,
  })
}

/**
 * Add a watch for environments in a specific namespace
 */
export async function watchNamespaceEnvironments(namespace: string): Promise<void> {
  await startWatch({
    group: 'environments.kloudlite.io',
    version: 'v1',
    plural: 'environments',
    cacheKeyPrefix: 'environments',
    namespace,
  })
}

/**
 * Check if watchers are active
 */
export function areWatchersActive(): boolean {
  return watchInitialized && activeWatches.size > 0
}

/**
 * Get watcher status
 */
export function getWatcherStatus() {
  return {
    initialized: watchInitialized,
    activeWatches: Array.from(activeWatches.entries()).map(([key, active]) => ({
      key,
      active,
    })),
  }
}

import { getK8sClient } from '@kloudlite/lib/k8s'
import { resourceStore } from './resource-store'

/**
 * K8s Resource Watcher — List-Then-Watch Pattern
 *
 * Populates the in-memory ResourceStore by:
 * 1. LIST all resources -> populate store -> markReady()
 * 2. WATCH from LIST's resourceVersion -> applyResource/removeResource on events
 * 3. On disconnect: re-LIST + re-WATCH (handles 410 Gone)
 *
 * Reads never hit the K8s API — they go straight to the ResourceStore.
 */

interface WatchConfig {
  group: string
  version: string
  plural: string
  namespace?: string
}

interface ListResponseLike {
  items?: unknown[]
  metadata?: {
    resourceVersion?: string
  }
}

function toListResponse(value: unknown): ListResponseLike {
  if (value && typeof value === 'object') {
    return value as ListResponseLike
  }
  return {}
}

// Survive HMR via globalThis
const g = globalThis as typeof globalThis & {
  __activeWatches?: Map<string, boolean>
  __watchInitialized?: boolean
  __namespaceWatchTimers?: Map<string, ReturnType<typeof setTimeout>>
}
const activeWatches = g.__activeWatches ??= new Map<string, boolean>()
let watchInitialized = g.__watchInitialized ?? false

// Track idle timers for namespace watches (5 min idle -> stop watches)
const NAMESPACE_IDLE_TIMEOUT = 5 * 60 * 1000
const namespaceWatchTimers = g.__namespaceWatchTimers ??= new Map<string, ReturnType<typeof setTimeout>>()

// Cluster-scoped resources to watch globally
const CLUSTER_WATCH_CONFIGS: WatchConfig[] = [
  { group: 'machines.kloudlite.io', version: 'v1', plural: 'machinetypes' },
  { group: 'machines.kloudlite.io', version: 'v1', plural: 'workmachines' },
  { group: 'platform.kloudlite.io', version: 'v1alpha1', plural: 'users' },
  { group: 'platform.kloudlite.io', version: 'v1alpha1', plural: 'userpreferences' },
]

// All namespace-scoped resource types
const NAMESPACE_WATCH_CONFIGS: Omit<WatchConfig, 'namespace'>[] = [
  { group: 'workspaces.kloudlite.io', version: 'v1', plural: 'workspaces' },
  { group: 'environments.kloudlite.io', version: 'v1', plural: 'environments' },
  { group: 'snapshots.kloudlite.io', version: 'v1', plural: 'snapshots' },
  { group: 'packages.kloudlite.io', version: 'v1', plural: 'packagerequests' },
  { group: '', version: 'v1', plural: 'services' },
  { group: '', version: 'v1', plural: 'configmaps' },
  { group: '', version: 'v1', plural: 'secrets' },
]

/**
 * Build the API watch/list path for a resource type
 */
function buildPath(config: WatchConfig): string {
  if (config.group) {
    return config.namespace
      ? `/apis/${config.group}/${config.version}/namespaces/${config.namespace}/${config.plural}`
      : `/apis/${config.group}/${config.version}/${config.plural}`
  }
  return config.namespace
    ? `/api/${config.version}/namespaces/${config.namespace}/${config.plural}`
    : `/api/${config.version}/${config.plural}`
}

/**
 * Perform an initial LIST to populate the store, returns resourceVersion for watch.
 */
async function performInitialList(config: WatchConfig): Promise<string> {
  const client = getK8sClient()
  let response: unknown

  if (config.group) {
    // CRD resources
    if (config.namespace) {
      response = await client.custom.listNamespacedCustomObject({
        group: config.group,
        version: config.version,
        namespace: config.namespace,
        plural: config.plural,
      })
    } else {
      response = await client.custom.listClusterCustomObject({
        group: config.group,
        version: config.version,
        plural: config.plural,
      })
    }
  } else {
    // Core API resources
    if (config.namespace) {
      switch (config.plural) {
        case 'services':
          response = await client.core.listNamespacedService({ namespace: config.namespace })
          break
        case 'configmaps':
          response = await client.core.listNamespacedConfigMap({ namespace: config.namespace })
          break
        case 'secrets':
          response = await client.core.listNamespacedSecret({ namespace: config.namespace })
          break
        default:
          throw new Error(`Unknown core resource: ${config.plural}`)
      }
    } else {
      throw new Error(`Cluster-scoped core resources not supported: ${config.plural}`)
    }
  }

  // Clear existing store data for this type+namespace before repopulating
  resourceStore.clearNamespace(config.plural, config.namespace)

  // Populate the store
  const listResponse = toListResponse(response)
  const items = listResponse.items || []
  for (const item of items) {
    resourceStore.applyResource(config.plural, item)
  }

  // Return resourceVersion for watch continuation
  const rv = listResponse.metadata?.resourceVersion || ''
  return rv
}

/**
 * Start a list-then-watch for a specific resource type.
 */
async function startWatch(config: WatchConfig): Promise<void> {
  const watchKey = `${config.group || 'core'}/${config.plural}/${config.namespace || 'cluster'}`

  if (activeWatches.get(watchKey)) {
    return // Already watching
  }

  console.log(`[K8S-WATCHER] Starting list-then-watch: ${watchKey}`)
  activeWatches.set(watchKey, true)

  try {
    // Phase 1: Initial LIST
    const resourceVersion = await performInitialList(config)
    resourceStore.markReady(config.plural, config.namespace)
    console.log(`[K8S-WATCHER] LIST complete for ${watchKey} (rv=${resourceVersion}, ${resourceStore.list(config.plural, config.namespace || '__cluster__').length} items)`)

    // Phase 2: WATCH from resourceVersion
    const client = getK8sClient()
    const k8sWatch = client.createWatch()
    const path = buildPath(config)

    await k8sWatch.watch(
      path,
      { resourceVersion },
      (type: string, obj: unknown) => {
        const resource = obj as { metadata?: { name?: string; namespace?: string } }
        const name = resource?.metadata?.name
        const namespace = resource?.metadata?.namespace

        if (type === 'ADDED' || type === 'MODIFIED') {
          resourceStore.applyResource(config.plural, resource)
        } else if (type === 'DELETED' && name) {
          resourceStore.removeResource(config.plural, namespace || '__cluster__', name)
        }

        console.log(`[K8S-WATCHER] ${config.plural} ${type}: ${name}`)
      },
      (err: unknown) => {
        const message = err instanceof Error ? err.message : String(err)
        console.error('[K8S-WATCHER] Watch error for %s: %s', watchKey, message)
        activeWatches.set(watchKey, false)
        resourceStore.markStale(config.plural, config.namespace)

        // Restart with full re-LIST after delay (handles 410 Gone)
        setTimeout(() => {
          console.log(`[K8S-WATCHER] Restarting: ${watchKey}`)
          startWatch(config).catch(console.error)
        }, 5000)
      },
    )
  } catch (err) {
    console.error(`[K8S-WATCHER] Failed to start ${watchKey}:`, err)
    activeWatches.set(watchKey, false)

    // Unblock any readers waiting on this resource type — they'll get
    // empty data, but the page renders instead of hanging forever
    resourceStore.markReady(config.plural, config.namespace)

    // Retry after delay
    setTimeout(() => {
      startWatch(config).catch(console.error)
    }, 5000)
  }
}

/**
 * Initialize watches for cluster-scoped resources.
 * Called once when the server starts (from instrumentation.ts).
 */
export async function initializeWatchers(): Promise<void> {
  if (watchInitialized) {
    return
  }

  console.log('[K8S-WATCHER] Initializing watchers...')
  watchInitialized = true
  g.__watchInitialized = true

  for (const config of CLUSTER_WATCH_CONFIGS) {
    startWatch(config).catch(console.error)
  }

  // Block until initial LISTs complete so the first request gets data immediately.
  // If K8s API is down, waitForReady's 10s timeout ensures we don't block forever.
  await Promise.all(
    CLUSTER_WATCH_CONFIGS.map((c) => resourceStore.waitForReady(c.plural))
  )
  console.log('[K8S-WATCHER] All cluster-scoped resources ready')
}

/**
 * Ensure watches are running for all resource types in a namespace.
 * Starts watches if not already running. Does NOT block — watches
 * populate the store in the background. Callers should use
 * resourceStore.waitForReady() for the specific type(s) they need.
 */
export function watchNamespace(namespace: string): void {
  // Reset idle timer
  const existingTimer = namespaceWatchTimers.get(namespace)
  if (existingTimer) clearTimeout(existingTimer)
  namespaceWatchTimers.set(namespace, setTimeout(() => {
    stopNamespaceWatches(namespace)
  }, NAMESPACE_IDLE_TIMEOUT))

  for (const config of NAMESPACE_WATCH_CONFIGS) {
    const watchKey = `${config.group || 'core'}/${config.plural}/${namespace}`
    if (!activeWatches.get(watchKey)) {
      startWatch({ ...config, namespace }).catch(console.error)
    }
  }
}

/**
 * Ensure a single resource type watch is running in a namespace.
 * Useful for watching only 'services' in environment targetNamespaces
 * to avoid connection accumulation (Issue 3).
 */
export async function watchResourceInNamespace(plural: string, namespace: string): Promise<void> {
  // Reset idle timer for this namespace
  const existingTimer = namespaceWatchTimers.get(namespace)
  if (existingTimer) clearTimeout(existingTimer)
  namespaceWatchTimers.set(namespace, setTimeout(() => {
    stopNamespaceWatches(namespace)
  }, NAMESPACE_IDLE_TIMEOUT))

  const config = NAMESPACE_WATCH_CONFIGS.find((c) => c.plural === plural)
  if (!config) {
    console.warn(`[K8S-WATCHER] Unknown resource type: ${plural}`)
    return
  }

  const watchKey = `${config.group || 'core'}/${plural}/${namespace}`
  if (!activeWatches.get(watchKey)) {
    startWatch({ ...config, namespace }).catch(console.error)
  }

  await resourceStore.waitForReady(plural, namespace)
}

/**
 * Stop all watches for a namespace (called on idle timeout).
 */
function stopNamespaceWatches(namespace: string): void {
  console.log(`[K8S-WATCHER] Stopping idle watches for namespace: ${namespace}`)
  for (const config of NAMESPACE_WATCH_CONFIGS) {
    const watchKey = `${config.group || 'core'}/${config.plural}/${namespace}`
    activeWatches.set(watchKey, false)
  }
  namespaceWatchTimers.delete(namespace)
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
    storeStats: resourceStore.getStats(),
  }
}

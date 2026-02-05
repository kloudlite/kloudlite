/**
 * In-Memory K8s Resource Store
 *
 * Replaces the LRU cache with a real-time in-memory store populated by K8s watches.
 * All K8s resources are kept in memory and updated via watch events (ADDED/MODIFIED/DELETED).
 * Reads are synchronous and zero-latency. Writes go through K8s API and the store
 * is updated when the watch event arrives.
 *
 * Flow:
 *   Watch start -> LIST all resources -> store in memory
 *   Watch event -> update store directly (ADDED/MODIFIED/DELETED)
 *   Read -> return from memory instantly
 *   Write -> K8s API -> watch event updates store -> next read returns fresh data
 */

// --- Types ---

interface ResourceTypeConfig {
  group: string // e.g., 'workspaces.kloudlite.io' or '' for core
  version: string // e.g., 'v1'
  plural: string // e.g., 'workspaces'
  namespaced: boolean
}

interface NamespaceStore<T = any> {
  resources: Map<string, T> // name -> resource
  hashIndex: Map<string, string> // hash -> name (from kloudlite.io/hash label)
  labelIndex: Map<string, Set<string>> // "key=value" -> Set<name>
}

type WatchState = 'pending' | 'ready' | 'stale'

const CLUSTER_KEY = '__cluster__'

// --- ResourceStore Class ---

class ResourceStore {
  private stores = new Map<string, Map<string, NamespaceStore>>()
  private configs = new Map<string, ResourceTypeConfig>()
  private readyState = new Map<string, WatchState>() // key: "plural:namespace"
  private readyResolvers = new Map<string, () => void>()
  private readyPromises = new Map<string, Promise<void>>()

  // --- Registration ---

  registerType(config: ResourceTypeConfig): void {
    this.configs.set(config.plural, config)
  }

  getConfig(plural: string): ResourceTypeConfig | undefined {
    return this.configs.get(plural)
  }

  getAllConfigs(): ResourceTypeConfig[] {
    return Array.from(this.configs.values())
  }

  // --- Readiness ---

  private readyKey(plural: string, namespace?: string): string {
    return `${plural}:${namespace || CLUSTER_KEY}`
  }

  markReady(plural: string, namespace?: string): void {
    const key = this.readyKey(plural, namespace)
    this.readyState.set(key, 'ready')
    const resolver = this.readyResolvers.get(key)
    if (resolver) {
      resolver()
      this.readyResolvers.delete(key)
      this.readyPromises.delete(key)
    }
  }

  markStale(plural: string, namespace?: string): void {
    const key = this.readyKey(plural, namespace)
    this.readyState.set(key, 'stale')
  }

  isReady(plural: string, namespace?: string): boolean {
    const key = this.readyKey(plural, namespace)
    return this.readyState.get(key) === 'ready'
  }

  /**
   * Wait for a specific resource type to be ready (populated via initial LIST).
   * Returns immediately if already ready. Times out after the specified duration.
   */
  waitForReady(plural: string, namespace?: string, timeoutMs = 10000): Promise<void> {
    const key = this.readyKey(plural, namespace)

    if (this.readyState.get(key) === 'ready') {
      return Promise.resolve()
    }

    // Reuse existing promise if waiting
    const existing = this.readyPromises.get(key)
    if (existing) return existing

    const promise = new Promise<void>((resolve) => {
      this.readyResolvers.set(key, resolve)

      // Ensure we don't wait forever
      setTimeout(() => {
        if (this.readyState.get(key) !== 'ready') {
          console.warn(`[RESOURCE-STORE] Timeout waiting for ${key} to be ready`)
          resolve()
          this.readyResolvers.delete(key)
          this.readyPromises.delete(key)
        }
      }, timeoutMs)
    })

    this.readyPromises.set(key, promise)
    return promise
  }

  // --- Private Helpers ---

  private getOrCreateNamespaceStore(plural: string, namespace: string): NamespaceStore {
    let typeStore = this.stores.get(plural)
    if (!typeStore) {
      typeStore = new Map()
      this.stores.set(plural, typeStore)
    }

    let nsStore = typeStore.get(namespace)
    if (!nsStore) {
      nsStore = {
        resources: new Map(),
        hashIndex: new Map(),
        labelIndex: new Map(),
      }
      typeStore.set(namespace, nsStore)
    }

    return nsStore
  }

  private extractName(resource: any): string {
    return resource?.metadata?.name || ''
  }

  private extractNamespace(resource: any): string {
    return resource?.metadata?.namespace || CLUSTER_KEY
  }

  private extractLabels(resource: any): Record<string, string> {
    return resource?.metadata?.labels || {}
  }

  private removeFromIndexes(nsStore: NamespaceStore, name: string, oldResource: any): void {
    // Remove from hash index
    const oldHash = this.extractLabels(oldResource)['kloudlite.io/hash']
    if (oldHash) {
      nsStore.hashIndex.delete(oldHash)
    }

    // Remove from label index
    const oldLabels = this.extractLabels(oldResource)
    for (const [key, value] of Object.entries(oldLabels)) {
      const indexKey = `${key}=${value}`
      const names = nsStore.labelIndex.get(indexKey)
      if (names) {
        names.delete(name)
        if (names.size === 0) {
          nsStore.labelIndex.delete(indexKey)
        }
      }
    }
  }

  private addToIndexes(nsStore: NamespaceStore, name: string, resource: any): void {
    // Add to hash index
    const hash = this.extractLabels(resource)['kloudlite.io/hash']
    if (hash) {
      nsStore.hashIndex.set(hash, name)
    }

    // Add to label index
    const labels = this.extractLabels(resource)
    for (const [key, value] of Object.entries(labels)) {
      const indexKey = `${key}=${value}`
      let names = nsStore.labelIndex.get(indexKey)
      if (!names) {
        names = new Set()
        nsStore.labelIndex.set(indexKey, names)
      }
      names.add(name)
    }
  }

  // --- Write Methods (called by watch handler) ---

  /**
   * Handle ADDED or MODIFIED watch events.
   * Updates the resource in the store and maintains all indexes.
   */
  applyResource(plural: string, resource: any): void {
    const name = this.extractName(resource)
    const namespace = this.extractNamespace(resource)
    if (!name) return

    const nsStore = this.getOrCreateNamespaceStore(plural, namespace)

    // If modifying, remove old indexes first
    const existing = nsStore.resources.get(name)
    if (existing) {
      this.removeFromIndexes(nsStore, name, existing)
    }

    // Store as plain object (K8s client returns objects with non-plain prototypes
    // which can't be passed from Server Components to Client Components)
    nsStore.resources.set(name, JSON.parse(JSON.stringify(resource)))

    // Update indexes
    this.addToIndexes(nsStore, name, resource)
  }

  /**
   * Handle DELETED watch events.
   * Removes the resource from the store and all indexes.
   */
  removeResource(plural: string, namespace: string, name: string): void {
    const nsKey = namespace || CLUSTER_KEY
    const typeStore = this.stores.get(plural)
    if (!typeStore) return

    const nsStore = typeStore.get(nsKey)
    if (!nsStore) return

    const existing = nsStore.resources.get(name)
    if (existing) {
      this.removeFromIndexes(nsStore, name, existing)
      nsStore.resources.delete(name)
    }
  }

  /**
   * Clear all resources for a given type+namespace.
   * Used before re-populating from a full LIST on reconnect.
   */
  clearNamespace(plural: string, namespace?: string): void {
    const nsKey = namespace || CLUSTER_KEY
    const typeStore = this.stores.get(plural)
    if (!typeStore) return

    typeStore.delete(nsKey)
  }

  // --- Read Methods (synchronous, from memory) ---

  /**
   * Get a single namespaced resource by name.
   */
  get<T = any>(plural: string, namespace: string, name: string): T | null {
    const typeStore = this.stores.get(plural)
    if (!typeStore) return null

    const nsStore = typeStore.get(namespace)
    if (!nsStore) return null

    return (nsStore.resources.get(name) as T) || null
  }

  /**
   * Get a single cluster-scoped resource by name.
   */
  getCluster<T = any>(plural: string, name: string): T | null {
    return this.get<T>(plural, CLUSTER_KEY, name)
  }

  /**
   * List all resources of a type in a namespace.
   */
  list<T = any>(plural: string, namespace: string): T[] {
    const typeStore = this.stores.get(plural)
    if (!typeStore) return []

    const nsStore = typeStore.get(namespace)
    if (!nsStore) return []

    return Array.from(nsStore.resources.values()) as T[]
  }

  /**
   * List all cluster-scoped resources of a type.
   */
  listCluster<T = any>(plural: string): T[] {
    return this.list<T>(plural, CLUSTER_KEY)
  }

  /**
   * Lookup a namespaced resource by its kloudlite.io/hash label.
   */
  getByHash<T = any>(plural: string, namespace: string, hash: string): T | null {
    const typeStore = this.stores.get(plural)
    if (!typeStore) return null

    const nsStore = typeStore.get(namespace)
    if (!nsStore) return null

    const name = nsStore.hashIndex.get(hash)
    if (!name) return null

    return (nsStore.resources.get(name) as T) || null
  }

  /**
   * Filter namespaced resources by a single label key=value.
   */
  listByLabel<T = any>(plural: string, namespace: string, key: string, value: string): T[] {
    const typeStore = this.stores.get(plural)
    if (!typeStore) return []

    const nsStore = typeStore.get(namespace)
    if (!nsStore) return []

    const indexKey = `${key}=${value}`
    const names = nsStore.labelIndex.get(indexKey)
    if (!names) return []

    const results: T[] = []
    for (const name of names) {
      const resource = nsStore.resources.get(name)
      if (resource) results.push(resource as T)
    }
    return results
  }

  /**
   * Filter cluster-scoped resources by a single label key=value.
   */
  listClusterByLabel<T = any>(plural: string, key: string, value: string): T[] {
    return this.listByLabel<T>(plural, CLUSTER_KEY, key, value)
  }

  /**
   * Find a namespaced resource where a status field matches a value.
   * Useful for fallback lookups like status.hash.
   */
  findByStatusField<T = any>(
    plural: string,
    namespace: string,
    fieldPath: string,
    value: string,
  ): T | null {
    const items = this.list<any>(plural, namespace)
    const parts = fieldPath.split('.')
    for (const item of items) {
      let current = item
      for (const part of parts) {
        current = current?.[part]
      }
      if (current === value) return item as T
    }
    return null
  }

  // --- Debug ---

  getStats(): Record<string, any> {
    const stats: Record<string, any> = {}
    for (const [plural, typeStore] of this.stores) {
      stats[plural] = {}
      for (const [ns, nsStore] of typeStore) {
        stats[plural][ns] = {
          count: nsStore.resources.size,
          hashIndexSize: nsStore.hashIndex.size,
          labelIndexSize: nsStore.labelIndex.size,
        }
      }
    }
    return stats
  }
}

// --- Singleton (survives HMR via globalThis) ---

const globalStore = globalThis as typeof globalThis & { __resourceStore?: ResourceStore }
export const resourceStore = globalStore.__resourceStore ??= new ResourceStore()

// --- Register all resource types ---

// Cluster-scoped
resourceStore.registerType({
  group: 'machines.kloudlite.io',
  version: 'v1',
  plural: 'machinetypes',
  namespaced: false,
})
resourceStore.registerType({
  group: 'machines.kloudlite.io',
  version: 'v1',
  plural: 'workmachines',
  namespaced: false,
})
resourceStore.registerType({
  group: 'platform.kloudlite.io',
  version: 'v1alpha1',
  plural: 'users',
  namespaced: false,
})
resourceStore.registerType({
  group: 'platform.kloudlite.io',
  version: 'v1alpha1',
  plural: 'userpreferences',
  namespaced: false,
})

// Namespace-scoped
resourceStore.registerType({
  group: 'workspaces.kloudlite.io',
  version: 'v1',
  plural: 'workspaces',
  namespaced: true,
})
resourceStore.registerType({
  group: 'environments.kloudlite.io',
  version: 'v1',
  plural: 'environments',
  namespaced: true,
})
resourceStore.registerType({
  group: 'snapshots.kloudlite.io',
  version: 'v1',
  plural: 'snapshots',
  namespaced: true,
})
resourceStore.registerType({
  group: 'packages.kloudlite.io',
  version: 'v1',
  plural: 'packagerequests',
  namespaced: true,
})
resourceStore.registerType({
  group: '',
  version: 'v1',
  plural: 'services',
  namespaced: true,
})
resourceStore.registerType({
  group: '',
  version: 'v1',
  plural: 'configmaps',
  namespaced: true,
})
resourceStore.registerType({
  group: '',
  version: 'v1',
  plural: 'secrets',
  namespaced: true,
})

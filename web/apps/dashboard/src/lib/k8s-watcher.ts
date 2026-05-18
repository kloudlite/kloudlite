/**
 * Dashboard no longer starts Kubernetes watches.
 *
 * These no-op exports keep older server actions callable while the dashboard
 * moves from direct Kubernetes access to platform-api reads/writes.
 */

export async function initializeWatchers(): Promise<void> {}

export function watchNamespace(_namespace: string): void {}

export async function watchResourceInNamespace(_plural: string, _namespace: string): Promise<void> {}

export function stopNamespaceWatches(_namespace: string): void {}

export function getWatcherStats() {
  return {
    initialized: false,
    activeWatches: [],
    namespaceTimers: [],
    storeStats: {},
  }
}

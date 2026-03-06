import type { WorkMachine } from '@kloudlite/lib/k8s'
import { resourceStore } from '@/lib/resource-store'

/**
 * Get work machine for a user from the in-memory store.
 */
export function getWorkMachineForUser(username: string): WorkMachine | null {
  const machines = resourceStore.listClusterByLabel<WorkMachine>(
    'workmachines',
    'kloudlite.io/owned-by',
    username,
  )
  return machines[0] || null
}

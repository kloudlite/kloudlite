'use server'

import type { WorkMachine } from '@kloudlite/lib/k8s'
import { resourceStore } from '@/lib/resource-store'

/**
 * Get WorkMachine status
 */
export async function getWorkMachineStatus(name: string) {
  try {
    await resourceStore.waitForReady('workmachines')
    const workMachine = resourceStore.getCluster<WorkMachine>('workmachines', name)

    if (!workMachine) {
      return {
        success: false,
        error: 'WorkMachine not found',
      }
    }

    return {
      success: true,
      data: {
        name: workMachine.metadata.name,
        state: workMachine.status?.state || workMachine.spec.state,
        isReady: workMachine.status?.isReady ?? false,
        publicIP: workMachine.status?.publicIP,
        message: workMachine.status?.message,
        lastUpdated: new Date().toISOString(),
      },
    }
  } catch (err) {
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

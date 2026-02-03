'use server'

import { workMachineRepository } from '@kloudlite/lib/k8s'

/**
 * Get WorkMachine status
 */
export async function getWorkMachineStatus(name: string) {
  try {
    const workMachine = await workMachineRepository.get(name)

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

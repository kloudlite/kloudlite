'use server'

import { workMachineRepository } from '@kloudlite/lib/k8s'

/**
 * Get WorkMachine metrics (CPU, memory, network)
 * Used for polling WorkMachine resource usage
 */
export async function getWorkMachineMetrics(name: string) {
  try {
    const workMachine = await workMachineRepository.get('', name)

    return {
      success: true,
      data: {
        name: workMachine.metadata!.name,
        cpu: workMachine.status?.metrics?.cpu || 0,
        memory: workMachine.status?.metrics?.memory || 0,
        network: {
          rx: workMachine.status?.metrics?.network?.rx || 0,
          tx: workMachine.status?.metrics?.network?.tx || 0,
        },
        disk: workMachine.status?.metrics?.disk || 0,
        timestamp: new Date().toISOString(),
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

/**
 * Get WorkMachine status
 */
export async function getWorkMachineStatus(name: string) {
  try {
    const workMachine = await workMachineRepository.get('', name)

    return {
      success: true,
      data: {
        name: workMachine.metadata!.name,
        state: workMachine.status?.state || workMachine.spec.state,
        isReady: workMachine.status?.isReady ?? false,
        publicIP: workMachine.status?.publicIP,
        phase: workMachine.status?.phase || 'Unknown',
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

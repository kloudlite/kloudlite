'use server'

import { workMachineRepository } from '@kloudlite/lib/k8s'
import type { WorkMachine } from '@kloudlite/lib/k8s'
import { getSession } from '@/lib/get-session'
import { resourceStore } from '@/lib/resource-store'

/**
 * Get the current authenticated username
 */
async function getCurrentUsername(): Promise<string> {
  const session = await getSession()
  if (!session?.user?.username) {
    throw new Error('Not authenticated')
  }
  return session.user.username
}

/**
 * Get work machine for a user from the in-memory store
 */
function getWorkMachineForUser(username: string): WorkMachine | null {
  const machines = resourceStore.listClusterByLabel<WorkMachine>('workmachines', 'kloudlite.io/owned-by', username)
  return machines[0] || null
}

export async function getMyWorkMachine() {
  try {
    console.log('[STORE] getMyWorkMachine')
    const username = await getCurrentUsername()
    await resourceStore.waitForReady('workmachines')
    const data = getWorkMachineForUser(username)
    if (!data) {
      return {
        success: false,
        error: 'No work machine found',
      }
    }
    return { success: true, data }
  } catch (err) {
    const error = err instanceof Error ? err : new Error('Unknown error')
    // Don't log error if user simply doesn't have a work machine yet
    if (!error.message.includes('not found') && !error.message.includes('No work machine found')) {
      console.error('Get my work machine error:', err)
    }
    return {
      success: false,
      error: error.message,
    }
  }
}

export async function listAllWorkMachines() {
  try {
    console.log('[STORE] listAllWorkMachines')
    await resourceStore.waitForReady('workmachines')
    const items = resourceStore.listCluster<WorkMachine>('workmachines')
    return { success: true, data: items }
  } catch (err) {
    console.error('List work machines error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

export async function startMyWorkMachine() {
  try {
    const username = await getCurrentUsername()
    await resourceStore.waitForReady('workmachines')
    const workMachine = getWorkMachineForUser(username)
    if (!workMachine) {
      return {
        success: false,
        error: 'No work machine found',
      }
    }
    console.log('[K8S-API] startMyWorkMachine:', workMachine.metadata!.name!)
    const data = await workMachineRepository.start(workMachine.metadata!.name!)
    return { success: true, data }
  } catch (err) {
    console.error('Start work machine error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

export async function stopMyWorkMachine() {
  try {
    const username = await getCurrentUsername()
    await resourceStore.waitForReady('workmachines')
    const workMachine = getWorkMachineForUser(username)
    if (!workMachine) {
      return {
        success: false,
        error: 'No work machine found',
      }
    }
    console.log('[K8S-API] stopMyWorkMachine:', workMachine.metadata!.name!)
    const data = await workMachineRepository.stop(workMachine.metadata!.name!)
    return { success: true, data }
  } catch (err) {
    console.error('Stop work machine error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

export async function createMyWorkMachine(machineType: string, volumeSize?: number) {
  try {
    const username = await getCurrentUsername()

    const workMachine: WorkMachine = {
      apiVersion: 'machines.kloudlite.io/v1',
      kind: 'WorkMachine',
      metadata: {
        name: `wm-${username}`,
        labels: {
          'kloudlite.io/owned-by': username,
        },
      },
      spec: {
        displayName: `${username}'s Work Machine`,
        ownedBy: username,
        machineType,
        state: 'running',
        volumeSize: volumeSize || 50,
        targetNamespace: `user-${username}`,
        sshPublicKeys: [],
        autoShutdown: {
          enabled: true,
          idleThresholdMinutes: 30,
          checkIntervalMinutes: 5,
        },
      },
    }

    console.log('[K8S-API] createMyWorkMachine:', workMachine.metadata?.name)
    const data = await workMachineRepository.create(workMachine)
    return { success: true, data }
  } catch (err) {
    console.error('Create work machine error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Admin action: assign a machine type to a user
 * Creates a WorkMachine if none exists, or patches the existing one
 */
export async function adminAssignMachineType(username: string, machineType: string) {
  try {
    // Verify caller is admin or super-admin
    const session = await getSession()
    if (!session?.user?.roles) {
      return { success: false, error: 'Not authenticated' }
    }
    const roles = session.user.roles
    if (!roles.includes('admin') && !roles.includes('super-admin')) {
      return { success: false, error: 'Insufficient permissions' }
    }

    await resourceStore.waitForReady('workmachines')
    await resourceStore.waitForReady('machinetypes')
    const existing = getWorkMachineForUser(username)

    // Read tier-specific defaults from MachineType annotations
    const machineTypeResource = resourceStore.getCluster('machinetypes', machineType)
    const ann = machineTypeResource?.metadata?.annotations || {}
    const tierStorageGb = parseInt(ann['kloudlite.io/tier-storage-gb'] || '0', 10) || 50
    const tierSuspendMinutes = parseInt(ann['kloudlite.io/tier-suspend-minutes'] || '0', 10) || 30

    if (existing) {
      // Patch existing work machine's machineType, volumeSize, and autoShutdown
      console.log('[K8S-API] adminAssignMachineType patch:', existing.metadata!.name!, machineType)
      const data = await workMachineRepository.patch(existing.metadata!.name!, {
        spec: {
          machineType,
          volumeSize: tierStorageGb,
          autoShutdown: {
            enabled: true,
            idleThresholdMinutes: tierSuspendMinutes,
            checkIntervalMinutes: 5,
          },
        },
      })
      return { success: true, data }
    } else {
      // Create a new work machine for this user (stopped state — admin assigns, user starts)
      const workMachine: WorkMachine = {
        apiVersion: 'machines.kloudlite.io/v1',
        kind: 'WorkMachine',
        metadata: {
          name: `wm-${username}`,
          labels: {
            'kloudlite.io/owned-by': username,
          },
        },
        spec: {
          displayName: `${username}'s Work Machine`,
          ownedBy: username,
          machineType,
          state: 'stopped',
          volumeSize: tierStorageGb,
          targetNamespace: `user-${username}`,
          sshPublicKeys: [],
          autoShutdown: {
            enabled: true,
            idleThresholdMinutes: tierSuspendMinutes,
            checkIntervalMinutes: 5,
          },
        },
      }
      console.log('[K8S-API] adminAssignMachineType create:', workMachine.metadata?.name)
      const data = await workMachineRepository.create(workMachine)
      return { success: true, data }
    }
  } catch (err) {
    console.error('Admin assign machine type error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return { success: false, error: error.message }
  }
}

export async function updateMyWorkMachine(updateData: {
  machineType?: string
  sshPublicKeys?: string[]
  autoShutdown?: {
    enabled: boolean
    idleThresholdMinutes: number
  }
}) {
  try {
    const username = await getCurrentUsername()
    await resourceStore.waitForReady('workmachines')
    const workMachine = getWorkMachineForUser(username)
    if (!workMachine) {
      return {
        success: false,
        error: 'No work machine found',
      }
    }

    console.log('[K8S-API] updateMyWorkMachine:', workMachine.metadata!.name!)
    const data = await workMachineRepository.patch(workMachine.metadata!.name!, {
      spec: updateData,
    })
    return { success: true, data }
  } catch (err) {
    console.error('Update work machine error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

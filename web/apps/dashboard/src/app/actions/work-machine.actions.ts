'use server'

import { cache } from 'react'
import { workMachineRepository } from '@kloudlite/lib/k8s'
import type { WorkMachine } from '@kloudlite/lib/k8s'
import { getSession } from '@/lib/get-session'
import { cachedFetch, CacheTTL, invalidateCache } from '@/lib/cache'

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
 * Cached work machine fetch - React cache for request deduplication + LRU for cross-request
 * Uses SHORT TTL (30s) because status (running/stopped) can change
 */
const getCachedWorkMachine = cache(async (username: string) => {
  return cachedFetch(
    `workMachine:${username}`,
    () => workMachineRepository.getByOwner(username),
    CacheTTL.SHORT // 30 seconds - status changes
  )
})

export async function getMyWorkMachine() {
  try {
    const username = await getCurrentUsername()
    const data = await getCachedWorkMachine(username)
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
    const result = await workMachineRepository.list('')
    return { success: true, data: result.items }
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
    console.log('[startMyWorkMachine] Starting...')
    const username = await getCurrentUsername()
    console.log('[startMyWorkMachine] Username:', username)
    const workMachine = await workMachineRepository.getByOwner(username)
    console.log('[startMyWorkMachine] Work machine found:', workMachine?.metadata?.name)
    if (!workMachine) {
      return {
        success: false,
        error: 'No work machine found',
      }
    }
    console.log('[startMyWorkMachine] Calling repository.start()...')
    const data = await workMachineRepository.start(workMachine.metadata!.name!)
    invalidateCache(`workMachine:${username}*`)
    console.log('[startMyWorkMachine] Success!')
    return { success: true, data }
  } catch (err) {
    console.error('[startMyWorkMachine] Error:', err)
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
    const workMachine = await workMachineRepository.getByOwner(username)
    if (!workMachine) {
      return {
        success: false,
        error: 'No work machine found',
      }
    }
    const data = await workMachineRepository.stop(workMachine.metadata!.name!)
    invalidateCache(`workMachine:${username}*`)
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

    const data = await workMachineRepository.create(workMachine)
    invalidateCache(`workMachine:${username}*`)
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
    const workMachine = await workMachineRepository.getByOwner(username)
    if (!workMachine) {
      return {
        success: false,
        error: 'No work machine found',
      }
    }

    // Use read-modify-write pattern for updates
    // Merge the update data into the existing spec
    const updatedMachine = {
      ...workMachine,
      spec: {
        ...workMachine.spec,
        ...updateData,
        // Deep merge autoShutdown if provided
        ...(updateData.autoShutdown && {
          autoShutdown: {
            ...workMachine.spec?.autoShutdown,
            ...updateData.autoShutdown,
          },
        }),
      },
    }

    // Cast to any to handle type differences between lib and types packages
    const data = await workMachineRepository.update(workMachine.metadata!.name!, updatedMachine as any)
    invalidateCache(`workMachine:${username}*`)
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

'use server'

import { getSession } from '@/lib/get-session'
import { getPlatformApiClient, type KubernetesResource } from '@/lib/platform-api'

type WorkMachine = KubernetesResource<{
  displayName?: string
  ownedBy?: string
  machineType?: string
  state?: string
  volumeSize?: number
  targetNamespace?: string
  sshPublicKeys?: string[]
  autoShutdown?: {
    enabled: boolean
    idleThresholdMinutes: number
    checkIntervalMinutes?: number
  }
}>

type MachineType = KubernetesResource<{
  displayName?: string
  resources?: { cpu?: string; memory?: string; gpu?: string }
}>

async function getCurrentUsername(): Promise<string> {
  const session = await getSession()
  if (!session?.user?.username) {
    throw new Error('Not authenticated')
  }
  return session.user.username
}

function getWorkMachineForUser(machines: WorkMachine[], username: string): WorkMachine | null {
  return machines.find((machine) => machine.metadata?.labels?.['kloudlite.io/owned-by'] === username || machine.spec?.ownedBy === username) || null
}

async function listWorkMachines() {
  const client = await getPlatformApiClient()
  return client.listClusterResources<WorkMachine>('workmachines')
}

export async function getMyWorkMachine() {
  try {
    const username = await getCurrentUsername()
    const data = getWorkMachineForUser(await listWorkMachines(), username)
    if (!data) return { success: false, error: 'No work machine found' }
    return { success: true, data }
  } catch (err) {
    const error = err instanceof Error ? err : new Error('Unknown error')
    return { success: false, error: error.message }
  }
}

export async function listAllWorkMachines() {
  try {
    return { success: true, data: await listWorkMachines() }
  } catch (err) {
    const error = err instanceof Error ? err : new Error('Unknown error')
    return { success: false, error: error.message }
  }
}

export async function startMyWorkMachine() {
  return patchMyWorkMachine({ state: 'running' })
}

export async function stopMyWorkMachine() {
  return patchMyWorkMachine({ state: 'stopped' })
}

export async function createMyWorkMachine(machineType: string, volumeSize?: number) {
  try {
    const username = await getCurrentUsername()
    const client = await getPlatformApiClient()
    const workMachine: WorkMachine = buildWorkMachine(username, machineType, 'running', volumeSize || 50, 30)
    const data = await client.createClusterResource<WorkMachine>('workmachines', workMachine)
    return { success: true, data }
  } catch (err) {
    const error = err instanceof Error ? err : new Error('Unknown error')
    return { success: false, error: error.message }
  }
}

export async function adminAssignMachineType(username: string, machineType: string) {
  try {
    const session = await getSession()
    const roles = session?.user?.roles || []
    if (!roles.includes('admin') && !roles.includes('super-admin')) {
      return { success: false, error: 'Insufficient permissions' }
    }

    const client = await getPlatformApiClient()
    const [machines, machineTypes] = await Promise.all([
      client.listClusterResources<WorkMachine>('workmachines'),
      client.listClusterResources<MachineType>('machinetypes'),
    ])
    const existing = getWorkMachineForUser(machines, username)
    const machineTypeResource = machineTypes.find((item) => item.metadata?.name === machineType)
    const ann = machineTypeResource?.metadata?.annotations || {}
    const tierStorageGb = parseInt(ann['kloudlite.io/tier-storage-gb'] || '0', 10) || 50
    const tierSuspendMinutes = parseInt(ann['kloudlite.io/tier-suspend-minutes'] || '0', 10) || 30

    if (existing?.metadata?.name) {
      const data = await client.patchClusterResource<WorkMachine>('workmachines', existing.metadata.name, {
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
    }

    const data = await client.createClusterResource<WorkMachine>('workmachines', buildWorkMachine(username, machineType, 'stopped', tierStorageGb, tierSuspendMinutes))
    return { success: true, data }
  } catch (err) {
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
  return patchMyWorkMachine(updateData)
}

async function patchMyWorkMachine(updateData: Record<string, unknown>) {
  try {
    const username = await getCurrentUsername()
    const workMachine = getWorkMachineForUser(await listWorkMachines(), username)
    const name = workMachine?.metadata?.name
    if (!name) return { success: false, error: 'No work machine found' }
    const client = await getPlatformApiClient()
    const data = await client.patchClusterResource<WorkMachine>('workmachines', name, { spec: updateData })
    return { success: true, data }
  } catch (err) {
    const error = err instanceof Error ? err : new Error('Unknown error')
    return { success: false, error: error.message }
  }
}

function buildWorkMachine(username: string, machineType: string, state: string, volumeSize: number, idleThresholdMinutes: number): WorkMachine {
  return {
    apiVersion: 'machines.kloudlite.io/v1',
    kind: 'WorkMachine',
    metadata: {
      name: `wm-${username}`,
      labels: { 'kloudlite.io/owned-by': username },
    },
    spec: {
      displayName: `${username}'s Work Machine`,
      ownedBy: username,
      machineType,
      state,
      volumeSize,
      targetNamespace: `user-${username}`,
      sshPublicKeys: [],
      autoShutdown: {
        enabled: true,
        idleThresholdMinutes,
        checkIntervalMinutes: 5,
      },
    },
  }
}

'use server'

import { revalidatePath } from 'next/cache'
import type { MachineTypeCreateRequest, MachineTypeUpdateRequest } from '@kloudlite/types'
import type { MachineType as DashboardMachineType } from '@/types/machine'
import { getPlatformApiClient, type KubernetesResource } from '@/lib/platform-api'

type MachineType = DashboardMachineType & KubernetesResource<DashboardMachineType['spec']>

/**
 * Server action to list all machine types
 * Reads from in-memory store (populated by K8s watcher)
 */
export async function listMachineTypes() {
  try {
    console.log('[STORE] listMachineTypes')
    const client = await getPlatformApiClient()
    const items = await client.listClusterResources<MachineType>('machinetypes')
    return { success: true, data: items }
  } catch (err) {
    console.error('List machine types error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to get a specific machine type
 */
export async function getMachineType(name: string) {
  try {
    console.log('[STORE] getMachineType:', name)
    const client = await getPlatformApiClient()
    const result = await client.getClusterResource<MachineType>('machinetypes', name)
    return { success: true, data: result }
  } catch (err) {
    console.error('Get machine type error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to create a machine type
 */
export async function createMachineType(data: MachineTypeCreateRequest) {
  try {
    const { name, ...specData } = data
    const machineType: MachineType = {
      apiVersion: 'machines.kloudlite.io/v1',
      kind: 'MachineType',
      metadata: {
        name,
      },
      spec: {
        displayName: specData.displayName || name,
        description: specData.description,
        category: specData.category,
        resources: {
          cpu: `${specData.cpu}`,
          memory: `${specData.memory}`,
          gpu: specData.gpu ? `${specData.gpu}` : undefined,
        },
        active: specData.active ?? true,
        isDefault: false,
      },
    }

    console.log('[K8S-API] createMachineType:', name)
    const client = await getPlatformApiClient()
    const result = await client.createClusterResource<MachineType>('machinetypes', machineType)
    revalidatePath('/admin/machine-configs')
    return { success: true, data: result }
  } catch (err) {
    console.error('Create machine type error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to update a machine type
 */
export async function updateMachineType(name: string, data: MachineTypeUpdateRequest) {
  try {
    // Convert update request to spec format
    const specUpdate: {
      displayName?: string
      description?: string
      category?: string
      active?: boolean
      resources?: {
        cpu?: string
        memory?: string
        gpu?: string
      }
    } = {}
    if (data.displayName !== undefined) specUpdate.displayName = data.displayName
    if (data.description !== undefined) specUpdate.description = data.description
    if (data.category !== undefined) specUpdate.category = data.category
    if (data.active !== undefined) specUpdate.active = data.active
    if (data.cpu !== undefined || data.memory !== undefined || data.gpu !== undefined) {
      specUpdate.resources = {}
      if (data.cpu !== undefined) specUpdate.resources.cpu = `${data.cpu}`
      if (data.memory !== undefined) specUpdate.resources.memory = `${data.memory}`
      if (data.gpu !== undefined) specUpdate.resources.gpu = `${data.gpu}`
    }

    // Use patch for partial updates
    console.log('[K8S-API] updateMachineType:', name)
    const client = await getPlatformApiClient()
    const result = await client.patchClusterResource<MachineType>('machinetypes', name, {
      spec: specUpdate,
    })
    revalidatePath('/admin/machine-configs')
    return { success: true, data: result }
  } catch (err) {
    console.error('Update machine type error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to delete a machine type
 */
export async function deleteMachineType(name: string) {
  try {
    console.log('[K8S-API] deleteMachineType:', name)
    const client = await getPlatformApiClient()
    await client.deleteClusterResource('machinetypes', name)
    revalidatePath('/admin/machine-configs')
    return { success: true }
  } catch (err) {
    console.error('Delete machine type error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to activate a machine type
 */
export async function activateMachineType(name: string) {
  try {
    console.log('[K8S-API] activateMachineType:', name)
    const client = await getPlatformApiClient()
    const result = await client.patchClusterResource<MachineType>('machinetypes', name, { spec: { active: true } })
    revalidatePath('/admin/machine-configs')
    return { success: true, data: result }
  } catch (err) {
    console.error('Activate machine type error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to deactivate a machine type
 */
export async function deactivateMachineType(name: string) {
  try {
    console.log('[K8S-API] deactivateMachineType:', name)
    const client = await getPlatformApiClient()
    const result = await client.patchClusterResource<MachineType>('machinetypes', name, { spec: { active: false } })
    revalidatePath('/admin/machine-configs')
    return { success: true, data: result }
  } catch (err) {
    console.error('Deactivate machine type error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to set a machine type as default
 */
export async function setMachineTypeAsDefault(name: string) {
  try {
    console.log('[K8S-API] setMachineTypeAsDefault:', name)
    const client = await getPlatformApiClient()
    const result = await client.patchClusterResource<MachineType>('machinetypes', name, { spec: { isDefault: true } })
    revalidatePath('/admin/machine-configs')
    revalidatePath('/')
    return { success: true, data: result }
  } catch (err) {
    console.error('Set machine type as default error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

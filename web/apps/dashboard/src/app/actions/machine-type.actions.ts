'use server'

import { revalidatePath } from 'next/cache'
import { machineTypeRepository } from '@kloudlite/lib/k8s'
import type { MachineType } from '@kloudlite/lib/k8s'
import type { MachineTypeCreateRequest, MachineTypeUpdateRequest } from '@kloudlite/types'
import { cachedFetch, CacheTTL, invalidateCache } from '@/lib/cache'

/**
 * Server action to list all machine types
 * Cached since machine types rarely change and this is called on every page load
 */
export async function listMachineTypes() {
  try {
    const result = await cachedFetch(
      'machineTypes:all',
      () => machineTypeRepository.list(),
      CacheTTL.STATIC // 5 minutes - machine types rarely change
    )
    return { success: true, data: result.items }
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
    const result = await machineTypeRepository.get(name)
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

    const result = await machineTypeRepository.create(machineType)
    invalidateCache('machineTypes:*')
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
    const specUpdate: any = {}
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
    const result = await machineTypeRepository.patch(name, {
      spec: specUpdate,
    })
    invalidateCache('machineTypes:*')
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
    await machineTypeRepository.delete(name)
    invalidateCache('machineTypes:*')
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
    const result = await machineTypeRepository.activate(name)
    invalidateCache('machineTypes:*')
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
    const result = await machineTypeRepository.deactivate(name)
    invalidateCache('machineTypes:*')
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
    const result = await machineTypeRepository.setDefault(name)
    invalidateCache('machineTypes:*')
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

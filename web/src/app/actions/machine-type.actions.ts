'use server'

import { revalidatePath } from 'next/cache'
import { machineTypeService } from '@/lib/services/machine-type.service'
import type { MachineTypeCreateRequest, MachineTypeUpdateRequest } from '@/types/machine'

/**
 * Server action to list all machine types
 */
export async function listMachineTypes() {
  try {
    const result = await machineTypeService.listMachineTypes()
    return { success: true, data: result }
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
    const result = await machineTypeService.getMachineType(name)
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
    const result = await machineTypeService.createMachineType(data)
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
    const result = await machineTypeService.updateMachineType(name, data)
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
    const result = await machineTypeService.deleteMachineType(name)
    revalidatePath('/admin/machine-configs')
    return { success: true, data: result }
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
    const result = await machineTypeService.activateMachineType(name)
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
    const result = await machineTypeService.deactivateMachineType(name)
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
    const result = await machineTypeService.setMachineTypeAsDefault(name)
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

'use server'

import { revalidatePath } from 'next/cache'
import { auth } from '@/lib/auth'
import { machineTypeService } from '@/lib/services/machine-type.service'
import type {
  MachineTypeCreateRequest,
  MachineTypeUpdateRequest
} from '@/types/machine'

/**
 * Server action to list all machine types
 */
export async function listMachineTypes() {
  try {
    // Get the current session
    const session = await auth()
    const userEmail = session?.user?.email

    const result = await machineTypeService.listMachineTypes(userEmail)
    return { success: true, data: result }
  } catch (error: any) {
    console.error('List machine types error:', error)
    return {
      success: false,
      error: error.message || 'Failed to list machine types'
    }
  }
}

/**
 * Server action to get a specific machine type
 */
export async function getMachineType(name: string) {
  try {
    // Get the current session
    const session = await auth()
    const userEmail = session?.user?.email

    const result = await machineTypeService.getMachineType(name, userEmail)
    return { success: true, data: result }
  } catch (error: any) {
    console.error('Get machine type error:', error)
    return {
      success: false,
      error: error.message || 'Failed to get machine type'
    }
  }
}

/**
 * Server action to create a machine type
 */
export async function createMachineType(data: MachineTypeCreateRequest) {
  try {
    // Get the current session
    const session = await auth()
    const userEmail = session?.user?.email

    const result = await machineTypeService.createMachineType(data, userEmail)
    revalidatePath('/admin/machine-configs')
    return { success: true, data: result }
  } catch (error: any) {
    console.error('Create machine type error:', error)
    return {
      success: false,
      error: error.message || 'Failed to create machine type'
    }
  }
}

/**
 * Server action to update a machine type
 */
export async function updateMachineType(name: string, data: MachineTypeUpdateRequest) {
  try {
    // Get the current session
    const session = await auth()
    const userEmail = session?.user?.email

    const result = await machineTypeService.updateMachineType(name, data, userEmail)
    revalidatePath('/admin/machine-configs')
    return { success: true, data: result }
  } catch (error: any) {
    console.error('Update machine type error:', error)
    return {
      success: false,
      error: error.message || 'Failed to update machine type'
    }
  }
}

/**
 * Server action to delete a machine type
 */
export async function deleteMachineType(name: string) {
  try {
    // Get the current session
    const session = await auth()
    const userEmail = session?.user?.email

    const result = await machineTypeService.deleteMachineType(name, userEmail)
    revalidatePath('/admin/machine-configs')
    return { success: true, data: result }
  } catch (error: any) {
    console.error('Delete machine type error:', error)
    return {
      success: false,
      error: error.message || 'Failed to delete machine type'
    }
  }
}

/**
 * Server action to activate a machine type
 */
export async function activateMachineType(name: string) {
  try {
    // Get the current session
    const session = await auth()
    const userEmail = session?.user?.email

    const result = await machineTypeService.activateMachineType(name, userEmail)
    revalidatePath('/admin/machine-configs')
    return { success: true, data: result }
  } catch (error: any) {
    console.error('Activate machine type error:', error)
    return {
      success: false,
      error: error.message || 'Failed to activate machine type'
    }
  }
}

/**
 * Server action to deactivate a machine type
 */
export async function deactivateMachineType(name: string) {
  try {
    // Get the current session
    const session = await auth()
    const userEmail = session?.user?.email

    const result = await machineTypeService.deactivateMachineType(name, userEmail)
    revalidatePath('/admin/machine-configs')
    return { success: true, data: result }
  } catch (error: any) {
    console.error('Deactivate machine type error:', error)
    return {
      success: false,
      error: error.message || 'Failed to deactivate machine type'
    }
  }
}
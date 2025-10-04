'use server'

import { revalidatePath } from 'next/cache'
import { environmentService } from '@/lib/services/environment.service'
import type {
  EnvironmentCreateRequest,
  EnvironmentUpdateRequest
} from '@/types/environment'

/**
 * Server action to create an environment
 */
export async function createEnvironment(data: EnvironmentCreateRequest) {
  try {
    const result = await environmentService.createEnvironment(data)
    revalidatePath('/environments')
    return { success: true, data: result }
  } catch (error: any) {
    console.error('Create environment error:', error)
    return {
      success: false,
      error: error.message || 'Failed to create environment'
    }
  }
}

/**
 * Server action to update an environment
 */
export async function updateEnvironment(name: string, data: EnvironmentUpdateRequest) {
  try {
    const result = await environmentService.updateEnvironment(name, data)
    revalidatePath('/environments')
    revalidatePath(`/environments/${name}`)
    return { success: true, data: result }
  } catch (error: any) {
    console.error('Update environment error:', error)
    return {
      success: false,
      error: error.message || 'Failed to update environment'
    }
  }
}

/**
 * Server action to delete an environment
 */
export async function deleteEnvironment(name: string) {
  try {
    const result = await environmentService.deleteEnvironment(name)
    revalidatePath('/environments')
    return { success: true, data: result }
  } catch (error: any) {
    console.error('Delete environment error:', error)
    return {
      success: false,
      error: error.message || 'Failed to delete environment'
    }
  }
}

/**
 * Server action to activate an environment
 */
export async function activateEnvironment(name: string) {
  try {
    const result = await environmentService.activateEnvironment(name)
    revalidatePath('/environments')
    revalidatePath(`/environments/${name}`)
    return { success: true, data: result }
  } catch (error: any) {
    console.error('Activate environment error:', error)
    return {
      success: false,
      error: error.message || 'Failed to activate environment'
    }
  }
}

/**
 * Server action to deactivate an environment
 */
export async function deactivateEnvironment(name: string) {
  try {
    const result = await environmentService.deactivateEnvironment(name)
    revalidatePath('/environments')
    revalidatePath(`/environments/${name}`)
    return { success: true, data: result }
  } catch (error: any) {
    console.error('Deactivate environment error:', error)
    return {
      success: false,
      error: error.message || 'Failed to deactivate environment'
    }
  }
}

/**
 * Server action to get environment status
 */
export async function getEnvironmentStatus(name: string) {
  try {
    const result = await environmentService.getEnvironmentStatus(name)
    return { success: true, data: result }
  } catch (error: any) {
    console.error('Get environment status error:', error)
    return {
      success: false,
      error: error.message || 'Failed to get environment status'
    }
  }
}
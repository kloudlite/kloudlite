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
export async function createEnvironment(data: EnvironmentCreateRequest, user: string) {
  try {
    const result = await environmentService.createEnvironment(data, user)
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
export async function updateEnvironment(name: string, data: EnvironmentUpdateRequest, user: string) {
  try {
    const result = await environmentService.updateEnvironment(name, data, user)
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
export async function deleteEnvironment(name: string, user: string) {
  try {
    const result = await environmentService.deleteEnvironment(name, user)
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
export async function activateEnvironment(name: string, user: string) {
  try {
    const result = await environmentService.activateEnvironment(name, user)
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
export async function deactivateEnvironment(name: string, user: string) {
  try {
    const result = await environmentService.deactivateEnvironment(name, user)
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
export async function getEnvironmentStatus(name: string, user: string) {
  try {
    const result = await environmentService.getEnvironmentStatus(name, user)
    return { success: true, data: result }
  } catch (error: any) {
    console.error('Get environment status error:', error)
    return {
      success: false,
      error: error.message || 'Failed to get environment status'
    }
  }
}
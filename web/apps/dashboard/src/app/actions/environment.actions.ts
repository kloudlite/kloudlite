'use server'

import { revalidatePath } from 'next/cache'
import { environmentService } from '@/lib/services/environment.service'
import type { EnvironmentCreateRequest, EnvironmentUpdateRequest } from '@kloudlite/types'

/**
 * Server action to create an environment
 */
export async function createEnvironment(data: EnvironmentCreateRequest) {
  try {
    const result = await environmentService.createEnvironment(data)
    revalidatePath('/environments')
    return { success: true, data: result }
  } catch (err) {
    console.error('Create environment error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
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
  } catch (err) {
    console.error('Update environment error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
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
  } catch (err) {
    console.error('Delete environment error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
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
  } catch (err) {
    console.error('Activate environment error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
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
  } catch (err) {
    console.error('Deactivate environment error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
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
  } catch (err) {
    console.error('Get environment status error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to clone an environment
 */
export async function cloneEnvironment(
  sourceName: string,
  targetName: string,
  targetNamespace: string,
  cloneEnvVars: boolean,
  cloneFiles: boolean,
  currentUser: string,
) {
  try {
    const result = await environmentService.cloneEnvironment(
      sourceName,
      targetName,
      targetNamespace,
      cloneEnvVars,
      cloneFiles,
      currentUser,
    )
    revalidatePath('/environments')
    return { success: true, data: result }
  } catch (err) {
    console.error('Clone environment error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

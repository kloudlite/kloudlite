'use server'

import { revalidatePath } from 'next/cache'
import { compositionService } from '@/lib/services/composition.service'
import type { CompositionCreateRequest, CompositionUpdateRequest } from '@/types/composition'

/**
 * Server action to create a composition
 */
export async function createComposition(namespace: string, data: CompositionCreateRequest) {
  try {
    const result = await compositionService.createComposition(namespace, data)
    revalidatePath(`/environments/${namespace}/resources/compositions`)
    return { success: true, data: result }
  } catch (err) {
    console.error('Create composition error:', err)
    const error = err instanceof Error ? err : new Error('Failed to create composition')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to update a composition
 */
export async function updateComposition(
  namespace: string,
  name: string,
  data: CompositionUpdateRequest,
) {
  try {
    const result = await compositionService.updateComposition(namespace, name, data)
    revalidatePath(`/environments/${namespace}/resources/compositions`)
    return { success: true, data: result }
  } catch (err) {
    console.error('Update composition error:', err)
    const error = err instanceof Error ? err : new Error('Failed to update composition')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to delete a composition
 */
export async function deleteComposition(namespace: string, name: string) {
  try {
    const result = await compositionService.deleteComposition(namespace, name)
    revalidatePath(`/environments/${namespace}/resources/compositions`)
    return { success: true, data: result }
  } catch (err) {
    console.error('Delete composition error:', err)
    const error = err instanceof Error ? err : new Error('Failed to delete composition')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to get composition status
 */
export async function getCompositionStatus(namespace: string, name: string) {
  try {
    const result = await compositionService.getCompositionStatus(namespace, name)
    return { success: true, data: result }
  } catch (err) {
    console.error('Get composition status error:', err)
    const error = err instanceof Error ? err : new Error('Failed to get composition status')
    return {
      success: false,
      error: error.message,
    }
  }
}

'use server'

import { revalidatePath } from 'next/cache'
import { registryService } from '@/lib/services/registry.service'

/**
 * Server action to list tags for a repository
 */
export async function listTags(repository: string) {
  try {
    const result = await registryService.listTags(repository)
    return { success: true, data: result }
  } catch (err) {
    console.error('List tags error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
      data: { name: repository, tags: [] },
    }
  }
}

/**
 * Server action to delete a tag from a repository
 */
export async function deleteTag(repository: string, tag: string) {
  try {
    await registryService.deleteTag(repository, tag)
    revalidatePath('/artifacts/container-images')
    return { success: true }
  } catch (err) {
    console.error('Delete tag error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to delete an entire repository (all tags)
 */
export async function deleteRepository(repository: string) {
  try {
    const result = await registryService.deleteRepository(repository)
    revalidatePath('/artifacts/container-images')
    return { success: true, data: result }
  } catch (err) {
    console.error('Delete repository error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

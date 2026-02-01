'use server'

import { revalidatePath } from 'next/cache'
import { userPreferencesRepository } from '@kloudlite/lib/k8s'
import { getSession } from '@/lib/get-session'

/**
 * Get the current authenticated username
 */
async function getCurrentUsername(): Promise<string> {
  const session = await getSession()
  if (!session?.user?.username) {
    throw new Error('Not authenticated')
  }
  return session.user.username
}

export async function getMyPreferences() {
  try {
    const username = await getCurrentUsername()
    const result = await userPreferencesRepository.getOrCreate('', username)
    return { success: true, data: result }
  } catch (err) {
    console.error('Get preferences error:', err)
    return { success: false, error: err instanceof Error ? err.message : 'Unknown error' }
  }
}

export async function pinWorkspace(name: string, namespace: string) {
  try {
    const username = await getCurrentUsername()
    await userPreferencesRepository.addPinnedWorkspace('', username, { name, namespace })
    revalidatePath('/') // Revalidate dashboard
    revalidatePath('/workspaces')
    return { success: true }
  } catch (err) {
    console.error('Pin workspace error:', err)
    return { success: false, error: err instanceof Error ? err.message : 'Unknown error' }
  }
}

export async function unpinWorkspace(name: string, namespace: string) {
  try {
    const username = await getCurrentUsername()
    await userPreferencesRepository.removePinnedWorkspace('', username, { name, namespace })
    revalidatePath('/')
    revalidatePath('/workspaces')
    return { success: true }
  } catch (err) {
    console.error('Unpin workspace error:', err)
    return { success: false, error: err instanceof Error ? err.message : 'Unknown error' }
  }
}

export async function pinEnvironment(name: string) {
  try {
    const username = await getCurrentUsername()
    await userPreferencesRepository.addPinnedEnvironment('', username, { name })
    revalidatePath('/')
    revalidatePath('/environments')
    return { success: true }
  } catch (err) {
    console.error('Pin environment error:', err)
    return { success: false, error: err instanceof Error ? err.message : 'Unknown error' }
  }
}

export async function unpinEnvironment(name: string) {
  try {
    const username = await getCurrentUsername()
    await userPreferencesRepository.removePinnedEnvironment('', username, { name })
    revalidatePath('/')
    revalidatePath('/environments')
    return { success: true }
  } catch (err) {
    console.error('Unpin environment error:', err)
    return { success: false, error: err instanceof Error ? err.message : 'Unknown error' }
  }
}

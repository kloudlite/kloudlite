'use server'

import { revalidatePath } from 'next/cache'
import { userPreferencesRepository } from '@kloudlite/lib/k8s'
import type { UserPreferences } from '@kloudlite/lib/k8s'
import { getSession } from '@/lib/get-session'
import { resourceStore } from '@/lib/resource-store'

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

    // Try store first
    await resourceStore.waitForReady('userpreferences')
    const prefs = resourceStore.getCluster<UserPreferences>('userpreferences', username)
    if (prefs) {
      return { success: true, data: prefs }
    }

    // Fall back to getOrCreate (creates the resource if it doesn't exist)
    const result = await userPreferencesRepository.getOrCreate(username)
    return { success: true, data: result }
  } catch (err) {
    console.error('Get preferences error:', err)
    return { success: false, error: err instanceof Error ? err.message : 'Unknown error' }
  }
}

export async function pinWorkspace(name: string, namespace: string) {
  try {
    const username = await getCurrentUsername()
    await userPreferencesRepository.addPinnedWorkspace(username, { name, namespace })
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
    await userPreferencesRepository.removePinnedWorkspace(username, { name, namespace })
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
    await userPreferencesRepository.addPinnedEnvironment(username, name)
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
    await userPreferencesRepository.removePinnedEnvironment(username, name)
    revalidatePath('/')
    revalidatePath('/environments')
    return { success: true }
  } catch (err) {
    console.error('Unpin environment error:', err)
    return { success: false, error: err instanceof Error ? err.message : 'Unknown error' }
  }
}

'use server'

import { revalidatePath } from 'next/cache'
import { userPreferencesService } from '@/lib/services/user-preferences.service'

export async function getMyPreferences() {
  try {
    const result = await userPreferencesService.getMyPreferences()
    return { success: true, data: result }
  } catch (err) {
    console.error('Get preferences error:', err)
    return { success: false, error: err instanceof Error ? err.message : 'Unknown error' }
  }
}

export async function pinWorkspace(name: string, namespace: string) {
  try {
    await userPreferencesService.pinWorkspace({ name, namespace })
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
    await userPreferencesService.unpinWorkspace({ name, namespace })
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
    await userPreferencesService.pinEnvironment({ name })
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
    await userPreferencesService.unpinEnvironment({ name })
    revalidatePath('/')
    revalidatePath('/environments')
    return { success: true }
  } catch (err) {
    console.error('Unpin environment error:', err)
    return { success: false, error: err instanceof Error ? err.message : 'Unknown error' }
  }
}

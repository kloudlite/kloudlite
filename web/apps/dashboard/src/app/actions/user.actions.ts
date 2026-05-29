'use server'

import { revalidatePath } from 'next/cache'
import { getPlatformApiClient, type KubernetesResource } from '@/lib/platform-api'

type User = KubernetesResource<{
  email?: string
  displayName?: string
  roles?: string[]
  active?: boolean
  passwordString?: string
  lastLoginAt?: string
}>

export interface ProviderAccount {
  provider: string
  providerId: string
  email: string
  name?: string | null
  image?: string | null
  connectedAt: string
}

export interface UserData {
  email: string
  name?: string | null
  image?: string | null
  provider: string
  providerId: string
}

/**
 * Get user by email (scan through store since there's no email label)
 */
export async function getUserByEmail(email: string) {
  try {
    console.log('[STORE] getUserByEmail:', email)
    const client = await getPlatformApiClient()
    const users = await client.listClusterResources<User>('users')
    const user = users.find((u) => u.spec?.email === email) || null
    if (!user) {
      return { success: false, error: 'User not found' }
    }
    return { success: true, data: user }
  } catch (err) {
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Update user's last login timestamp
 */
export async function updateUserLastLogin(username: string) {
  try {
    console.log('[K8S-API] updateUserLastLogin:', username)
    const client = await getPlatformApiClient()
    const updated = await client.patchClusterResource<User>('users', username, {
      spec: { lastLoginAt: new Date().toISOString() },
    })
    return { success: true, data: updated }
  } catch (err) {
    console.error('Update last login error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Reset user password
 */
export async function resetUserPassword(username: string, newPassword: string) {
  try {
    // Set passwordString - the mutation webhook will hash it with bcrypt
    // and store the result in spec.password
    const client = await getPlatformApiClient()
    const updated = await client.patchClusterResource<User>('users', username, {
      spec: {
        passwordString: newPassword,
      },
    })

    revalidatePath('/admin/users')
    return { success: true, data: updated }
  } catch (err) {
    console.error('Reset password error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Check if username is available
 */
export async function checkUsernameAvailability(username: string) {
  try {
    const client = await getPlatformApiClient()
    let user: User | null = null
    try {
      user = await client.getClusterResource<User>('users', username)
    } catch {
      user = null
    }
    return {
      success: true,
      data: {
        available: !user,
      },
    }
  } catch (err) {
    console.error('Check username availability error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * List all users
 */
export async function listUsers() {
  try {
    console.log('[STORE] listUsers')
    const client = await getPlatformApiClient()
    const items = await client.listClusterResources<User>('users')
    return { success: true, data: items }
  } catch (err) {
    console.error('List users error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

// Alias for compatibility
export const getAllUsers = listUsers

/**
 * Get user by username
 */
export async function getUser(username: string) {
  try {
    console.log('[STORE] getUser:', username)
    const client = await getPlatformApiClient()
    const user = await client.getClusterResource<User>('users', username)
    if (!user) {
      return { success: false, error: 'User not found' }
    }
    return { success: true, data: user }
  } catch (err) {
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Create user
 */
export async function createUser(userData: {
  username: string
  email: string
  displayName?: string
  password?: string
  roles: string[]
  isActive?: boolean
}) {
  try {
    const user = {
      apiVersion: 'platform.kloudlite.io/v1alpha1',
      kind: 'User',
      metadata: {
        name: userData.username,
      },
      spec: {
        email: userData.email,
        displayName: userData.displayName || userData.username,
        roles: userData.roles,
        active: userData.isActive ?? true,
        ...(userData.password ? { passwordString: userData.password } : {}),
      },
    }

    console.log('[PLATFORM-API] createUser:', userData.username)
    const client = await getPlatformApiClient()
    const created = await client.createClusterResource<User>('users', user)
    revalidatePath('/admin/users')
    return { success: true, data: created }
  } catch (err) {
    console.error('Create user error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Update user
 */
export async function updateUser(
  username: string,
  updates: {
    email?: string
    displayName?: string
    roles?: string[]
    isActive?: boolean
  }
) {
  try {
    console.log('[PLATFORM-API] updateUser:', username)
    const client = await getPlatformApiClient()
    const updated = await client.patchClusterResource<User>('users', username, {
      spec: {
        email: updates.email,
        displayName: updates.displayName,
        roles: updates.roles,
        active: updates.isActive,
      },
    })

    revalidatePath('/admin/users')
    return { success: true, data: updated }
  } catch (err) {
    console.error('Update user error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Delete user
 */
export async function deleteUser(username: string) {
  try {
    console.log('[PLATFORM-API] deleteUser:', username)
    const client = await getPlatformApiClient()
    await client.deleteClusterResource('users', username)
    revalidatePath('/admin/users')
    return { success: true }
  } catch (err) {
    console.error('Delete user error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Activate user
 */
export async function activateUser(username: string) {
  try {
    console.log('[K8S-API] activateUser:', username)
    const client = await getPlatformApiClient()
    const updated = await client.patchClusterResource<User>('users', username, { spec: { active: true } })
    revalidatePath('/admin/users')
    return { success: true, data: updated }
  } catch (err) {
    console.error('Activate user error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Deactivate user
 */
export async function deactivateUser(username: string) {
  try {
    console.log('[K8S-API] deactivateUser:', username)
    const client = await getPlatformApiClient()
    const updated = await client.patchClusterResource<User>('users', username, { spec: { active: false } })
    revalidatePath('/admin/users')
    return { success: true, data: updated }
  } catch (err) {
    console.error('Deactivate user error:', err)
    const error = err instanceof Error ? err : new Error('Unknown error')
    return {
      success: false,
      error: error.message,
    }
  }
}

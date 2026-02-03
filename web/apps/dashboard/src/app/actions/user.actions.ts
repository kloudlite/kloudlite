'use server'

import { revalidatePath } from 'next/cache'
import { userRepository } from '@kloudlite/lib/k8s'
import bcrypt from 'bcryptjs'

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
 * Get user by email
 */
export async function getUserByEmail(email: string) {
  try {
    const user = await userRepository.getByEmail(email)
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
    const updated = await userRepository.updateLastLogin(username)
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
    // Hash the new password
    const salt = await bcrypt.genSalt(10)
    const hashedPassword = await bcrypt.hash(newPassword, salt)

    // Encode to base64 for storage
    const passwordHashBase64 = Buffer.from(hashedPassword).toString('base64')

    // Update user with new password hash
    const updated = await userRepository.patch(username, {
      spec: {
        passwordHash: passwordHashBase64,
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
    // Try to get user by name
    try {
      await userRepository.get(username)
      // User exists, username not available
      return {
        success: true,
        data: {
          available: false,
        },
      }
    } catch (err) {
      // User not found, username available
      return {
        success: true,
        data: {
          available: true,
        },
      }
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
    const result = await userRepository.list()
    return { success: true, data: result.items }
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
    const user = await userRepository.get(username)
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
  roles: string[]
  password: string
  isActive?: boolean
}) {
  try {
    // Hash password
    const salt = await bcrypt.genSalt(10)
    const hashedPassword = await bcrypt.hash(userData.password, salt)
    const passwordHashBase64 = Buffer.from(hashedPassword).toString('base64')

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
        passwordHash: passwordHashBase64,
      },
    }

    const created = await userRepository.create(user as any)
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
    const updated = await userRepository.patch(username, {
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
    await userRepository.delete(username)
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
    const updated = await userRepository.activate(username)
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
    const updated = await userRepository.deactivate(username)
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

'use server'

import { apiClient } from '@/lib/api-client'
import { userService, type User, type CreateUserRequest, type UpdateUserRequest } from '@/lib/services/user.service'
import { userToDisplay, type UserDisplay, type CreateUserFormData, type UserResource } from '@/types/user'
import { revalidatePath } from 'next/cache'

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

export async function authenticateUser(userData: UserData) {
  try {
    // Get user by email using efficient endpoint
    let existingUser
    try {
      existingUser = await apiClient.get<UserResource>(`/api/v1/users/by-email?email=${encodeURIComponent(userData.email)}`)
    } catch (err) {
      const error = err instanceof Error ? err : new Error('Unknown error')
      // Check if it's a 404 from our API (user not found)
      if (error.message?.includes('404')) {
        // User doesn't exist, authentication should fail
        console.log(`Authentication failed: User with email ${userData.email} not found`)
        return {
          success: false,
          error: `You are not registered. Please contact your administrator to create an account for ${userData.email}.`
        }
      }
      // For other errors, return a generic message
      console.error('Error during authentication:', error)
      return {
        success: false,
        error: 'Authentication failed. Please try again later.'
      }
    }

    // User exists, update their provider information and last login
    const providerAccount: ProviderAccount = {
      provider: userData.provider,
      providerId: userData.providerId,
      email: userData.email,
      name: userData.name,
      image: userData.image,
      connectedAt: new Date().toISOString(),
    }

    // Check if this provider is already connected
    const existingProviders = existingUser.spec.providers || []
    const providerIndex = existingProviders.findIndex(
      (p: ProviderAccount) => p.provider === userData.provider
    )

    let updatedProviders: ProviderAccount[]
    if (providerIndex >= 0) {
      // Update existing provider info
      updatedProviders = [...existingProviders]
      updatedProviders[providerIndex] = providerAccount
    } else {
      // Add new provider to existing user
      updatedProviders = [...existingProviders, providerAccount]
    }

    // Update user with latest provider info
    const updatePayload = {
      ...existingUser.spec,
      providers: updatedProviders,
      metadata: {
        ...(existingUser.spec.metadata ?? {}),
        lastProvider: userData.provider,
        totalProviders: updatedProviders.length.toString(),
      }
    }

    // Update using the Kubernetes resource name from metadata
    const updatedUser = await apiClient.put(
      `/api/v1/users/${existingUser.metadata.name}`,
      updatePayload
    )

    // Update last login in status field using dedicated endpoint
    try {
      await apiClient.post(`/api/v1/users/${existingUser.metadata.name}/update-last-login`)
    } catch (error) {
      console.warn('Failed to update last login status:', error)
      // Don't fail authentication if this fails
    }

    console.log(`User ${userData.email} authenticated successfully with provider: ${userData.provider}`)
    return {
      success: true,
      user: updatedUser,
      message: 'Authentication successful'
    }

  } catch (error) {
    console.error('Error during authentication:', error)
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Authentication failed'
    }
  }
}

export async function checkUserExists(email: string): Promise<boolean> {
  try {
    await apiClient.get<UserResource>(`/api/v1/users/by-email?email=${encodeURIComponent(email)}`)
    return true
  } catch (err) {
    // Type guard for error objects with response property
    if (err && typeof err === 'object' && 'response' in err) {
      const errorWithResponse = err as { response?: { status?: number } }
      if (errorWithResponse.response?.status === 404) {
        return false
      }
    }
    console.error('Error checking user existence:', err)
    return false
  }
}

// User management CRUD operations
export async function getAllUsers(): Promise<{ success: boolean; users?: UserDisplay[]; error?: string }> {
  try {
    const users = await userService.listUsers()
    const displayUsers = users.map(userToDisplay)

    return {
      success: true,
      users: displayUsers
    }
  } catch (error) {
    console.error('Error fetching users:', error)
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Failed to fetch users'
    }
  }
}

export async function createUser(data: CreateUserFormData): Promise<{ success: boolean; user?: UserDisplay; error?: string }> {
  try {
    const createData: CreateUserRequest = {
      email: data.email,
      displayName: data.displayName,
      roles: data.roles,
      isActive: true
    }

    const user = await userService.createUser(createData)

    // Revalidate the users page to show updated data
    revalidatePath('/admin/users')

    return {
      success: true,
      user: userToDisplay(user)
    }
  } catch (error) {
    console.error('Error creating user:', error)
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Failed to create user'
    }
  }
}

export async function updateUser(userName: string, data: Partial<CreateUserFormData & { isActive?: boolean }>): Promise<{ success: boolean; user?: UserDisplay; error?: string }> {
  try {
    // If only updating active status, use domain-specific endpoints
    if (data.isActive !== undefined && Object.keys(data).length === 1) {
      const user = data.isActive
        ? await userService.activateUser(userName)
        : await userService.deactivateUser(userName)

      // Revalidate the users page to show updated data
      revalidatePath('/admin/users')

      return {
        success: true,
        user: userToDisplay(user)
      }
    }

    // For other fields, use the full update
    const updateData: UpdateUserRequest = {
      email: data.email,
      displayName: data.displayName,
      roles: data.roles,
      isActive: data.isActive
    }

    const user = await userService.updateUser(userName, updateData)

    // Revalidate the users page to show updated data
    revalidatePath('/admin/users')

    return {
      success: true,
      user: userToDisplay(user)
    }
  } catch (error) {
    console.error('Error updating user:', error)
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Failed to update user'
    }
  }
}

export async function deleteUser(userName: string): Promise<{ success: boolean; error?: string }> {
  try {
    await userService.deleteUser(userName)

    // Revalidate the users page to show updated data
    revalidatePath('/admin/users')

    return {
      success: true
    }
  } catch (error) {
    console.error('Error deleting user:', error)
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Failed to delete user'
    }
  }
}

export async function resetUserPassword(userName: string, newPassword: string): Promise<{ success: boolean; error?: string }> {
  try {
    await apiClient.post(`/api/v1/users/${userName}/reset-password`, {
      newPassword: newPassword
    })

    // Revalidate the users page
    revalidatePath('/admin/users')

    return {
      success: true
    }
  } catch (error) {
    console.error('Error resetting user password:', error)
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Failed to reset password'
    }
  }
}


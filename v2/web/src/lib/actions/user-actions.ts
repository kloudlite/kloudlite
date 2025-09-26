'use server'

import { apiClient } from '@/lib/api-client'

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
      existingUser = await apiClient.get<any>(`/api/v1/users/by-email?email=${encodeURIComponent(userData.email)}`)
    } catch (error: any) {
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

    // Update user with latest provider info and login time
    const updatePayload = {
      ...existingUser.spec,
      providers: updatedProviders,
      metadata: {
        ...(existingUser.spec.metadata ?? {}),
        lastLoginAt: new Date().toISOString(),
        lastProvider: userData.provider,
        totalProviders: updatedProviders.length.toString(),
      }
    }

    // Update using the Kubernetes resource name from metadata
    const updatedUser = await apiClient.put(
      `/api/v1/users/${existingUser.metadata.name}`,
      updatePayload
    )

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
    await apiClient.get<any>(`/api/v1/users/by-email?email=${encodeURIComponent(email)}`)
    return true
  } catch (error: any) {
    if (error.response?.status === 404) {
      return false
    }
    console.error('Error checking user existence:', error)
    return false
  }
}
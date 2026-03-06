'use server'

import { userRepository } from '@kloudlite/lib/k8s'
import type { User } from '@kloudlite/lib/k8s'
import { resourceStore } from '@/lib/resource-store'

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
 * Authenticate user with OAuth provider
 * Updates provider information and last login timestamp
 */
export async function authenticateUser(userData: UserData) {
  try {
    // Try store first, fall back to direct API call for reliability during auth
    console.log('[STORE] authenticateUser: looking up', userData.email)
    await resourceStore.waitForReady('users', undefined, 5000)
    const users = resourceStore.listCluster<User>('users')
    let user = users.find((u) => u.spec?.email === userData.email) || null

    if (!user) {
      // Fallback to direct API if store is empty or user not found
      try {
        console.log('[K8S-API] authenticateUser: fallback getByEmail', userData.email)
        user = await userRepository.getByEmail(userData.email)
      } catch (_err) {
        console.log(`Authentication failed: User with email ${userData.email} not found`)
        return {
          success: false,
          error: `You are not registered. Please contact your administrator to create an account for ${userData.email}.`,
        }
      }
    }

    // Create provider account info
    const providerAccount: ProviderAccount = {
      provider: userData.provider,
      providerId: userData.providerId,
      email: userData.email,
      name: userData.name,
      image: userData.image,
      connectedAt: new Date().toISOString(),
    }

    // Check if this provider is already connected
    const existingProviders = user.spec?.providers || []
    const providerIndex = existingProviders.findIndex((p) => p.provider === userData.provider)

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
    console.log('[K8S-API] authenticateUser: patch providers for', user.metadata!.name!)
    const updatedUser = await userRepository.patch(user.metadata!.name!, {
      spec: {
        providers: updatedProviders,
      },
    })

    // Update last login timestamp
    try {
      await userRepository.updateLastLogin(user.metadata!.name!)
    } catch (error) {
      console.warn('Failed to update last login status:', error)
      // Don't fail authentication if this fails
    }

    console.log(
      `User ${userData.email} authenticated successfully with provider: ${userData.provider}`,
    )

    return {
      success: true,
      user: updatedUser,
      message: 'Authentication successful',
    }
  } catch (error) {
    console.error('Error during authentication:', error)
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Authentication failed',
    }
  }
}

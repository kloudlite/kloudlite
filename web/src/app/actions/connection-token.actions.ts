'use server'

import { revalidatePath } from 'next/cache'
import { connectionTokenService } from '@/lib/services/connection-token.service'
import type { CreateConnectionTokenRequest } from '@/lib/services/connection-token.service'

/**
 * Server action to list connection tokens
 */
export async function listConnectionTokens() {
  try {
    const result = await connectionTokenService.listTokens()
    return { success: true, data: result }
  } catch (error: any) {
    console.error('List connection tokens error:', error)
    return {
      success: false,
      error: error.message || 'Failed to list connection tokens',
    }
  }
}

/**
 * Server action to create a connection token
 */
export async function createConnectionToken(data: CreateConnectionTokenRequest) {
  try {
    const result = await connectionTokenService.createToken(data)
    revalidatePath('/connection-tokens')
    return { success: true, data: result }
  } catch (error: any) {
    console.error('Create connection token error:', error)
    return {
      success: false,
      error: error.message || 'Failed to create connection token',
    }
  }
}

/**
 * Server action to delete a connection token
 */
export async function deleteConnectionToken(name: string) {
  try {
    await connectionTokenService.deleteToken(name)
    revalidatePath('/connection-tokens')
    return { success: true }
  } catch (error: any) {
    console.error('Delete connection token error:', error)
    return {
      success: false,
      error: error.message || 'Failed to delete connection token',
    }
  }
}

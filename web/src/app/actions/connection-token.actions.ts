'use server'

import { revalidatePath } from 'next/cache'
import { env } from '@/lib/env'
import { connectionTokenService } from '@/lib/services/connection-token.service'
import type { CreateConnectionTokenRequest } from '@/lib/services/connection-token.service'

/**
 * Server action to list connection tokens
 */
export async function listConnectionTokens() {
  try {
    const result = await connectionTokenService.listTokens()
    return { success: true, data: result }
  } catch (err) {
    console.error('List connection tokens error:', err)
    const error = err instanceof Error ? err : new Error("Unknown error")
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Server action to create a connection token
 */
export async function createConnectionToken(data: Omit<CreateConnectionTokenRequest, 'webUrl'>) {
  try {
    // Get the webUrl from environment variables
    const result = await connectionTokenService.createToken({
      ...data,
      webUrl: env.webUrl
    })
    revalidatePath('/connection-tokens')
    return { success: true, data: result }
  } catch (err) {
    console.error('Create connection token error:', err)
    const error = err instanceof Error ? err : new Error("Unknown error")
    return {
      success: false,
      error: error.message,
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
  } catch (err) {
    console.error('Delete connection token error:', err)
    const error = err instanceof Error ? err : new Error("Unknown error")
    return {
      success: false,
      error: error.message,
    }
  }
}

/**
 * Magic link token storage and management
 */

import { supabase } from '../supabase'
import type { Database } from '../supabase-types'
import { randomBytes } from 'crypto'

type MagicLinkTokenInsert = Database['public']['Tables']['magic_link_tokens']['Insert']
type MagicLinkTokenUpdate = Database['public']['Tables']['magic_link_tokens']['Update']

const TOKEN_EXPIRATION_MINUTES = 15

/**
 * Generate a cryptographically secure token
 * Uses 32 random bytes (256 bits) encoded as base64url
 * Results in a 43-character URL-safe string
 */
function generateSecureToken(): string {
  return randomBytes(32).toString('base64url')
}

/**
 * Create a new magic link token for an email address
 * @param email - User's email address
 * @param ipAddress - Optional IP address for audit trail
 * @param userAgent - Optional user agent for audit trail
 * @returns The generated token string
 */
export async function createMagicLinkToken(
  email: string,
  ipAddress?: string,
  userAgent?: string
): Promise<string> {
  const token = generateSecureToken()
  const expiresAt = new Date(Date.now() + TOKEN_EXPIRATION_MINUTES * 60 * 1000)

  const insertData: MagicLinkTokenInsert = {
    email,
    token,
    expires_at: expiresAt.toISOString(),
    ip_address: ipAddress || null,
    user_agent: userAgent || null,
  }

  const { error } = await supabase
    .from('magic_link_tokens')
    // @ts-expect-error - Supabase client with placeholder values has type issues during build
    .insert(insertData)

  if (error) {
    console.error('Failed to create magic link token:', error)
    throw new Error('Failed to create magic link token')
  }

  return token
}

/**
 * Verify a magic link token and return the associated email
 * Checks:
 * - Token exists
 * - Token has not expired
 * - Token has not been used
 * @param token - The token to verify
 * @returns The email address if valid, null otherwise
 */
export async function verifyMagicLinkToken(
  token: string
): Promise<string | null> {
  const { data, error } = await supabase
    .from('magic_link_tokens')
    .select('*')
    .eq('token', token)
    .maybeSingle()

  if (error || !data) {
    return null
  }

  // Check if token has already been used
  // @ts-expect-error - Supabase client with placeholder values has type issues during build
  if (data.used_at) {
    return null
  }

  // Check if token has expired
  // @ts-expect-error - Supabase client with placeholder values has type issues during build
  const expiresAt = new Date(data.expires_at)
  if (expiresAt < new Date()) {
    return null
  }

  // @ts-expect-error - Supabase client with placeholder values has type issues during build
  return data.email
}

/**
 * Mark a token as used to prevent reuse
 * @param token - The token to mark as used
 */
export async function markTokenAsUsed(token: string): Promise<void> {
  const updateData: MagicLinkTokenUpdate = {
    used_at: new Date().toISOString(),
  }

  const { error } = await supabase
    .from('magic_link_tokens')
    // @ts-expect-error - Supabase client with placeholder values has type issues during build
    .update(updateData)
    .eq('token', token)

  if (error) {
    console.error('Failed to mark token as used:', error)
    throw new Error('Failed to mark token as used')
  }
}

/**
 * Clean up expired tokens from the database
 * Should be called periodically (e.g., via cron job)
 */
export async function cleanupExpiredTokens(): Promise<void> {
  const { error } = await supabase
    .from('magic_link_tokens')
    .delete()
    .lt('expires_at', new Date().toISOString())

  if (error) {
    console.error('Failed to clean up expired tokens:', error)
  }
}
